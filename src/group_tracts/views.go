package group_tracts

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const RANGE_LIMIT int = 100

type scoresQuery struct {
	limit  int
	offset int
}

type Address struct {
	Address string `json:"address"`
}

type Vintage struct {
	IsDefault          bool   `json:"isDefault"`
	ID                 string `json:"id"`
	VintageName        string `json:"vintageName"`
	VintageDescription string `json:"vintageDescription"`
}

type Benchmark struct {
	IsDefault            bool   `json:"isDefault"`
	BenchmarkDescription string `json:"benchmarkDescription"`
	ID                   string `json:"id"`
	BenchmarkName        string `json:"benchmarkName"`
}

type Input struct {
	Address   Address   `json:"address"`
	Vintage   Vintage   `json:"vintage"`
	Benchmark Benchmark `json:"benchmark"`
}
type TigerLine struct {
	Side        string `json:"side"`
	TigerLineID string `json:"tigerLineId"`
}

type CensusBlock struct {
	Suffix    string `json:"SUFFIX"`
	Pop100    int    `json:"POP100"`
	Geoid     string `json:"GEOID"`
	Cenlat    string `json:"CENTLAT"`
	Block     string `json:"BLOCK"`
	Areawater int    `json:"AREAWATER"`
	State     string `json:"STATE"`
	Basename  string `json:"BASENAME"`
	OID       int
	LSADC     string
	INTPTLAT  string
	FUNCSTAT  string
	NAME      string
	OBJECTID  int
	TRACT     string
	CENTLON   string
	BLKGRP    string
	AREALAND  int
	HU100     int
	INTPTLON  string
	MTFCC     string
	LWBLKTYP  string
	UR        string
	COUNTY    string
}

type StateLDUpper []CensusBlock
type CensusBlocks []CensusBlock

type Geographies struct {
	CensusBlocks `json:"Census Blocks"`
}

type Coordinates struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type AddressComponents struct {
	Zip             string `json:"zip"`
	StreetName      string `json:"streetName"`
	PreType         string `json:"preType"`
	City            string `json:"city"`
	PreDirection    string `json:"preDirection"`
	SuffixDirection string `json:"suffixDirection"`
	FromAddress     string `json:"fromAddress"`
	State           string `json:"state"`
	SuffixType      string `json:"suffixType"`
	ToAddress       string `json:"toAddress"`
	SuffixQualifier string `json:"suffixQualifier"`
	PreQualifier    string `json:"preQualifier"`
}

type AddressMatch struct {
	TigerLine         TigerLine         `json:"tigerLine"`
	Geographies       Geographies       `json:"geographies"`
	Coordinates       Coordinates       `json:"coordinates"`
	AddressComponents AddressComponents `json:"addressComponents"`
	MatchedAddress    string            `json:"matchedAddress"`
}
type AddressMatches []AddressMatch

type GeoCodingResultDetail struct {
	Input          Input          `json:"input"`
	AddressMatches AddressMatches `json:"addressMatches"`
}

type GeoCodingResult struct {
	Result GeoCodingResultDetail `json:"result"`
}

type AddressScoreResult struct {
	Geoid                          uint64  `json:"geoid"`
	NWI                            float64 `json:"nwi"`
	RegionalTransitUsagePercentage float64 `json:"regionalTransitUsagePercentage"`
	RegionalBikeRidership          uint64  `json:"regionalBikeRidership"`
	SearchedAddress                string  `json:"searchedAddress"`
}

type ScoreResult struct {
	ID                             int     `json:"id"`
	Geoid                          uint64  `json:"geoid"`
	CSA_name                       string  `json:"csa_name"`
	CBSA_name                      string  `json:"cbsa_name"`
	NWI                            float64 `json:"nwi"`
	RegionalTransitUsagePercentage float64 `json:"regionalTransitUsagePercentage"`
	RegionalBikeRidership          uint64  `json:"regionalBikeRidership"`
}

type ScoreResults []ScoreResult

type ZipcodeResult struct {
	Zipcode string `json:"zipcode"`
}

func (h handler) GetScoreByAddress(ctx *gin.Context) {
	address := strings.ReplaceAll(ctx.Query("address"), " ", "%20")
	geoid := make(chan string)
	var results GeoCodingResult
	if address != "" {
		go func() {
			url := "https://geocoding.geo.census.gov/geocoder/geographies/onelineaddress?address=" + address + "&benchmark=2020&vintage=Census2010_Census2020&format=json"
			client := &http.Client{}
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}
			req.Header.Set("User-Agent", "Golang_Spider_Bot/3.0")
			resp, err := client.Do(req)
			if err != nil {
				ctx.AbortWithError(http.StatusBadRequest, err)
				return
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				ctx.AbortWithError(http.StatusBadRequest, err)
				return
			}
			if err := json.Unmarshal(body, &results); err != nil { // Parse []byte to go struct pointer
				ctx.AbortWithError(http.StatusBadRequest, err)
				return
			}
			if len(results.Result.AddressMatches) > 0 {
				geoids := results.Result.AddressMatches[0].Geographies.CensusBlocks[0].Geoid
				geoid <- geoids[:len(geoids)-3]
			} else {
				geoid <- ""
			}
		}()
	}
	var score Rank
	var cbsa CBSA
	geoid10, err := strconv.ParseUint(<-geoid, 10, 64)
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, err)
		return
	}
	if result := h.DB.Where(&Rank{Geoid: geoid10}).First(&score); result.Error != nil {
		ctx.AbortWithError(http.StatusNotFound, result.Error)
		return
	}
	if result := h.DB.Where(&CBSA{Geoid: geoid10}).First(&cbsa); result.Error != nil {
		ctx.AbortWithError(http.StatusNotFound, result.Error)
		return
	}
	result := AddressScoreResult{
		Geoid:                          geoid10,
		NWI:                            score.NWI,
		SearchedAddress:                ctx.Query("address"),
		RegionalTransitUsagePercentage: cbsa.PublicTansitUsage,
		RegionalBikeRidership:          cbsa.BikeRidership,
	}
	ctx.JSON(http.StatusOK, &result)
}

func (h handler) GetScores(ctx *gin.Context) {
	var scores []Rank
	var zipScores []Rank
	var res []Zipcode
	zipcode := ctx.Query("zipcode")
	limit, limit_err := strconv.Atoi(ctx.Query("limit"))
	if limit_err != nil {
		limit = 50
	}
	offset, offset_err := strconv.Atoi(ctx.Query("offset"))
	if offset_err != nil {
		offset = 0
	}
	query := scoresQuery{
		limit:  limit,
		offset: offset,
	}
	if zipcode == "" {
		if result := h.DB.Limit(query.limit).Offset(query.offset).Find(&scores); result.Error != nil {
			ctx.AbortWithError(http.StatusNotFound, result.Error)
			return
		}
	} else {
		offset := ctx.Query("offset")
		if offset != "" {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}
		if result := h.DB.Where(&Zipcode{Zipcode: zipcode}).Find(&res); result.Error != nil {
			ctx.AbortWithError(http.StatusNotFound, result.Error)
			return
		}
		for _, item := range res {
			if result := h.DB.Where(&CBSA{CBSA: item.CBSA}).Model(&Rank{}).Select("*").Joins("left join cbsas on cbsas.geoid = ranks.geoid").Limit(query.limit).Scan(&zipScores); result.Error != nil {
				fmt.Println(result.Error)
			}
			scores = append(scores, zipScores...)
		}
	}
	results := ScoreResults{}
	for i := range scores {
		var csa CSA
		var cbsa CBSA
		if csa_result := h.DB.Where(&CSA{Geoid: scores[i].Geoid}).First(&csa); csa_result.Error != nil {
			csa.CSA_name = ""
		}
		if cbsa_result := h.DB.Where(&CBSA{Geoid: scores[i].Geoid}).First(&cbsa); cbsa_result.Error != nil {
			cbsa.CBSA_name = ""
		}
		results = append(
			results,
			ScoreResult{
				ID:                             i + offset,
				Geoid:                          scores[i].Geoid,
				CSA_name:                       csa.CSA_name,
				CBSA_name:                      cbsa.CBSA_name,
				NWI:                            scores[i].NWI,
				RegionalTransitUsagePercentage: cbsa.PublicTansitUsage,
				RegionalBikeRidership:          cbsa.BikeRidership,
			},
		)
	}
	ctx.JSON(http.StatusOK, &results)
}
