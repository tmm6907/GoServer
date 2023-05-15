package group_tracts

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const RANGE_LIMIT int = 100
// func (h handler) GetScoreByAddress(ctx *gin.Context) {
// 	address := strings.ReplaceAll(ctx.Query("address"), " ", "%20")
// 	geoid := make(chan string)
// 	var results GeoCodingResult
// 	if address != "" {
		
// 	}
	
// }

func (h handler) GetScores(ctx *gin.Context) {
	address := strings.ReplaceAll(ctx.Query("address"), " ", "%20")
	if address != ""{
		var geoidResults GeoCodingResult
		geoid := make(chan string)
		go func() {
			url := "https://geocoding.geo.census.gov/geocoder/geographies/onelineaddress?address=" + address + "&benchmark=2020&vintage=Census2010_Census2020&format=json"
			client := &http.Client{}
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}
			log.Println("Request = ", req)
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
			if err := json.Unmarshal(body, &geoidResults); err != nil { // Parse []byte to go struct pointer
				ctx.AbortWithError(http.StatusBadRequest, err)
				return
			}
			if len(geoidResults.Result.AddressMatches) > 0 {
				geoids := geoidResults.Result.AddressMatches[0].Geographies.CensusBlocks[0].Geoid
				geoid <- geoids[:len(geoids)-3]
			} else {
				geoid <- ""
			}
		}()
		var score Rank
		var cbsa CBSA
		geoid10, err := strconv.ParseUint(<-geoid, 10, 64)
		log.Println(geoid10)
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
	}else{
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
		query := ScoresQuery{
			Limit:  limit,
			Offset: offset,
		}
		if zipcode == "" {
			if result := h.DB.Limit(query.limit).Offset(query.offset).Find(&scores); result.Error != nil {
				ctx.AbortWithError(http.StatusNotFound, result.Error)
				return
			}
		} else {
			if result := h.DB.Limit(query.limit).Offset(query.offset).Where("zipcode=?", zipcode).Find(&res); result.Error != nil {
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
			if csa_result := h.DB.Where(&CBSA{Geoid: scores[i].Geoid}).First(&csa); csa_result.Error != nil {
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
}
