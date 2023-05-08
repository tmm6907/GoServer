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

type Result struct {
	Input          Input          `json:"input"`
	AddressMatches AddressMatches `json:"addressMatches"`
}

type GeoCodingResult struct {
	Result Result `json:"result"`
}

func (h handler) GetScore(ctx *gin.Context) {
	id := ctx.Param("id")
	var group_tract GroupTract

	if result := h.DB.First(&group_tract, id); result.Error != nil {
		ctx.AbortWithError(http.StatusNotFound, result.Error)
		return
	}

	ctx.JSON(http.StatusOK, &group_tract)
}

func (h handler) GetScoreByAddress(ctx *gin.Context) {
	address := strings.ReplaceAll(ctx.Query("address"), " ", "%20")
	geoid := make(chan string)
	var results GeoCodingResult
	if address != "" {
		go func() {
			url := "https://geocoding.geo.census.gov/geocoder/geographies/onelineaddress?address=" + address + "&benchmark=2020&vintage=Census2010_Census2020&format=json"
			fmt.Println("URL: ", url)
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
				fmt.Println("Can not unmarshal JSON")
			}
			if len(results.Result.AddressMatches) > 0 {
				geoids := results.Result.AddressMatches[0].Geographies.CensusBlocks[0].Geoid
				geoid <- geoids[:len(geoids)-3]
			} else {
				geoid <- ""
			}
		}()
	}
	var group_tract GroupTract
	geoid10, err := strconv.ParseUint(<-geoid, 10, 64)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	fmt.Println(geoid10)
	if result := h.DB.Where(&GroupTract{Geoid10: geoid10}).First(&group_tract); result.Error != nil {
		ctx.AbortWithError(http.StatusNotFound, result.Error)
		return
	}
	ctx.JSON(http.StatusOK, &group_tract)
}

func (h handler) GetScores(ctx *gin.Context) {
	var scores []Rank
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
	if result := h.DB.Limit(query.limit).Offset(query.offset).Find(&scores); result.Error != nil {
		ctx.AbortWithError(http.StatusNotFound, result.Error)
		return
	}

	ctx.JSON(http.StatusOK, &scores)
}
