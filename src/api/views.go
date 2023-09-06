package api

import (
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"nwi.io/nwi/caches"
	"nwi.io/nwi/models"
	"nwi.io/nwi/serializers"
)

const (
	D2a                     string = "d2a"
	D2b                            = "d2b"
	D3b                            = "d3b"
	D4a                            = "d4a"
	D2aRanked                      = "d2aRanked"
	D2bRanked                      = "d2bRanked"
	D3bRanked                      = "d3bRanked"
	D4aRanked                      = "d4aRanked"
	PublicTransitEstimate          = "publicTransitEstimate"
	PublicTransitPercentage        = "publicTransitPercentage"
	BikeRidership                  = "bikeRidership"
	BikeRidershipPercentage        = "bikeRidershipPercentage"
)
const CONNECTION_LIMIT = 1000

func (h handler) GetScores(ctx *gin.Context) {
	var query serializers.ScoreQuery
	err := ctx.Bind(&query)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "bad request",
		})
		return
	}
	query.SetFormat()
	switch {
	case query.Address != "":
		geoid := make(chan string)
		go func() {
			log.Println("Fetching Geoid...")
			timerStart := time.Now()
			res, err := query.GetGeoid()
			timerEnd := time.Now()
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{
					"error": "no results found",
				})
				return
			}
			duration := timerEnd.Sub(timerStart).Milliseconds()
			log.Printf("Time elapsed %vms", duration)
			geoid <- res
		}()
		var score models.Rank
		geoid10, err := strconv.ParseUint(<-geoid, 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": "no results found",
			})
			return
		}
		if result := h.DB.Where(&models.Rank{Geoid: geoid10}).Select("ranks.id", "ranks.geoid", "ranks.transit_score", "ranks.bike_score", "ranks.nwi").First(&score); result.Error != nil {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": "no results found",
			})
			return
		}
		switch strings.ToLower(query.Format) {
		case "json":
			result := serializers.AddressScoreResult{
				RankID:          score.ID,
				Geoid:           geoid10,
				NWI:             score.NWI,
				TransitScore:    score.TransitScore,
				BikeScore:       score.BikeScore,
				SearchedAddress: query.Address,
				Format:          query.Format,
			}
			caches.CACHE.Put(ctx.Request.URL.RequestURI(), result)
			ctx.JSON(http.StatusOK, &result)
			return
		case "xml":
			result := serializers.AddressScoreResultXML{
				XMLName:         xml.Name{Space: "result"},
				RankID:          score.ID,
				Geoid:           geoid10,
				NWI:             score.NWI,
				TransitScore:    score.TransitScore,
				BikeScore:       score.BikeScore,
				SearchedAddress: query.Address,
				Format:          query.Format,
			}
			caches.CACHE.Put(ctx.Request.URL.RequestURI(), result)

			ctx.XML(http.StatusOK, &result)
			return
		default:
		}
	case query.ZipCode != "":
		var ranks []models.Rank
		var cbsas []uint32
		var cbsaName []string
		var geoids []uint64
		query.SetLimit()
		if result := h.DB.Model(models.ZipCode{}).Select("cbsa").Where(&models.ZipCode{Zipcode: query.ZipCode}).Find(&cbsas); result.Error != nil {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": "no results found",
			})
			return
		}
		if result := h.DB.Select("ranks.id", "ranks.geoid", "ranks.transit_score", "ranks.bike_score", "ranks.nwi").Limit(query.Limit).Offset(query.Offset).Where("cbsas.cbsa IN ?", cbsas).Model(&models.Rank{}).Joins("left join cbsas on cbsas.geoid = ranks.geoid").Scan(&ranks); result.Error != nil {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": "no results found",
			})
		}
		for _, rank := range ranks {
			geoids = append(geoids, rank.Geoid)
		}
		h.DB.Model(models.CBSA{}).Select("cbsa_name").Where("geoid IN ?", geoids).Find(&cbsaName)
		switch query.Format {
		case "json":
			var _wg sync.WaitGroup
			resultScores := make(chan serializers.ScoreResults, len(ranks))
			for i, rank := range ranks {
				_wg.Add(1)
				go func(i int, rank models.Rank) {
					defer _wg.Done()
					resultScores <- serializers.ScoreResults{
						RankID:       rank.ID,
						Geoid:        rank.Geoid,
						CBSAName:     cbsaName[i],
						NWI:          rank.NWI,
						TransitScore: rank.TransitScore,
						BikeScore:    rank.BikeScore,
						Format:       query.Format,
					}
				}(i, rank)
			}
			_wg.Wait()
			close(resultScores)
			results := []serializers.ScoreResults{}
			for result := range resultScores {
				results = append(results, result)
			}
			caches.CACHE.Put(ctx.Request.URL.RequestURI(), results)
			ctx.JSON(http.StatusOK, &results)
			return
		case "xml":
			var _wg sync.WaitGroup
			resultScores := make(chan serializers.ScoreResultsXML, len(ranks))
			for i, rank := range ranks {
				_wg.Add(1)
				go func(i int, rank models.Rank) {
					defer _wg.Done()
					resultScores <- serializers.ScoreResultsXML{
						XMLName:      xml.Name{Space: "result"},
						RankID:       rank.ID,
						Geoid:        rank.Geoid,
						CBSAName:     cbsaName[i],
						NWI:          rank.NWI,
						TransitScore: rank.TransitScore,
						BikeScore:    rank.BikeScore,
						Format:       query.Format,
					}
				}(i, rank)
			}
			_wg.Wait()
			close(resultScores)
			results := serializers.XMLResults{}
			for result := range resultScores {
				results.Scores = append(results.Scores, result)
			}
			caches.CACHE.Put(ctx.Request.URL.RequestURI(), results)
			ctx.XML(http.StatusOK, &results)
			return
		}
	default:
		ctx.AbortWithError(http.StatusBadRequest, errors.New("bad request! make sure all required fields are filled"))
		return
	}
}

func (h handler) GetScoreDetails(ctx *gin.Context) {
	var query serializers.DetailQuery
	var searchedRank models.Rank
	var cbsa models.CBSA
	stringRankID := ctx.Param("id")
	if stringRankID == "" {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("bad request! provide rankID as id parameter"))
		return
	}
	if err := ctx.Bind(&query); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("bad request! make sure you've provided valid parameters"))
		return
	}

	rankID, err := strconv.ParseUint(stringRankID, 10, 64)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("bad request! provide a valid rankID"))
		return
	}

	if rankResult := h.DB.First(&searchedRank, uint(rankID)); rankResult.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "no results found",
		})
		return
	}
	if cbsaResult := h.DB.First(&cbsa, searchedRank.Geoid); cbsaResult.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "no results found",
		})
		return
	}
	query.SetFormat()
	if query.Fields == "" {
		switch query.Format {
		case "json":
			result := serializers.DetailResult{
				RankID:                  searchedRank.ID,
				D2b_e8mixa:              searchedRank.D2b_e8mixa,
				D2a_ephhm:               searchedRank.D2a_ephhm,
				D3b:                     searchedRank.D3b,
				D4a:                     searchedRank.D4a,
				PublicTransitEstimate:   cbsa.PublicTransitEstimate,
				PublicTransitPercentage: cbsa.PublicTransitPercentage,
				BikeRidership:           cbsa.BikeRidership,
				BikeRidershipPercentage: cbsa.BikeRidershipPercentage,
				BikeRidershipRank:       searchedRank.BikeCountRank,
				BikePercentageRank:      searchedRank.BikePercentageRank,
				BikeFatalityRank:        searchedRank.BikeFatalityRank,
				BikeShareRank:           searchedRank.BikeShareRank,
			}
			caches.CACHE.Put(ctx.Request.URL.RequestURI(), result)
			ctx.JSON(http.StatusOK, &result)
			return
		case "xml":
			result := serializers.DetailResultXML{
				RankID:                  searchedRank.ID,
				D2b_e8mixa:              searchedRank.D2b_e8mixa,
				D2a_ephhm:               searchedRank.D2a_ephhm,
				D3b:                     searchedRank.D3b,
				D4a:                     searchedRank.D4a,
				PublicTransitEstimate:   cbsa.PublicTransitEstimate,
				PublicTransitPercentage: cbsa.PublicTransitPercentage,
				BikeRidership:           cbsa.BikeRidership,
				BikeRidershipPercentage: cbsa.BikeRidershipPercentage,
				BikeRidershipRank:       searchedRank.BikeCountRank,
				BikePercentageRank:      searchedRank.BikePercentageRank,
				BikeFatalityRank:        searchedRank.BikeFatalityRank,
				BikeShareRank:           searchedRank.BikeShareRank,
			}
			caches.CACHE.Put(ctx.Request.URL.RequestURI(), result)
			ctx.XML(http.StatusOK, &result)
			return
		}
	} else {
		result := make(map[string]interface{})
		result["rankID"] = rankID
		fields := strings.Split(query.Fields, ",")
		for _, field := range fields {
			switch field {
			case D2a:
				result["d2a"] = searchedRank.D2a_ephhm
			case D2b:
				result["d2b"] = searchedRank.D2b_e8mixa
			case D3b:
				result["d3b"] = searchedRank.D3b
			case D4a:
				result["d4a"] = searchedRank.D4a
			case D2aRanked:
				result["d2aRanked"] = searchedRank.D2a_ranked
			case D2bRanked:
				result["d2bRanked"] = searchedRank.D2b_ranked
			case D3bRanked:
				result["d3bRanked"] = searchedRank.D3b_ranked
			case D4aRanked:
				result["d4aRanked"] = searchedRank.D4a_ranked
			case PublicTransitEstimate:
				result["publicTransitEstimate"] = cbsa.PublicTransitEstimate
			case PublicTransitPercentage:
				result["publicTransitPercentage"] = cbsa.PublicTransitPercentage
			case BikeRidership:
				result["bikeRidership"] = cbsa.BikeRidership
			case BikeRidershipPercentage:
				result["bikeRidershipPercentage"] = cbsa.BikeRidershipPercentage
			default:
				ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid field name: '%s'", field))
				return
			}
		}
		switch query.Format {
		case "json":
			result["format"] = "json"
			caches.CACHE.Put(ctx.Request.URL.RequestURI(), result)
			ctx.JSON(http.StatusOK, &result)
			return
		case "xml":
			result["format"] = "xml"
			result["result"] = xml.Name{}
			caches.CACHE.Put(ctx.Request.URL.RequestURI(), result)
			ctx.XML(http.StatusOK, &result)
			return
		default:
		}

	}
}
