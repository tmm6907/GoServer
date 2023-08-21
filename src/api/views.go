package api

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	query.SetFormat()
	switch {
	case query.Address != "":
		geoid := make(chan string)
		go func() {
			res, err := query.GetGeoid()
			if err != nil {
				ctx.AbortWithError(http.StatusNotFound, err)
				return
			}
			geoid <- res
		}()
		var score models.Rank
		geoid10, err := strconv.ParseUint(<-geoid, 10, 64)
		if err != nil {
			ctx.AbortWithError(http.StatusNotFound, err)
			return
		}
		if result := h.DB.Where(&models.Rank{Geoid: geoid10}).First(&score); result.Error != nil {
			ctx.AbortWithError(http.StatusNotFound, result.Error)
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
		var zipScores []models.Rank
		var res []models.ZipCode
		var scores []models.Rank
		var _wg sync.WaitGroup
		workers := make(chan struct{}, CONNECTION_LIMIT)
		var mu sync.RWMutex
		query.SetLimit()
		if result := h.DB.Where(&models.ZipCode{Zipcode: query.ZipCode}).Find(&res); result.Error != nil {
			ctx.AbortWithError(http.StatusNotFound, result.Error)
			return
		}
		for _, item := range res {
			_wg.Add(1)
			go func(item models.ZipCode) {
				workers <- struct{}{}
				if result := h.DB.Limit(query.Limit).Offset(query.Offset).Where("cbsas.cbsa = ?", item.CBSA).Model(&models.Rank{}).Joins("left join cbsas on cbsas.geoid = ranks.geoid").Scan(&zipScores); result.Error != nil {
					fmt.Println(result.Error)
				}
				mu.Lock()
				scores = append(scores, zipScores...)
				mu.Unlock()
				defer func() {
					<-workers
					_wg.Done()
				}()
			}(item)
		}
		_wg.Wait()
		switch query.Format {
		case "json":
			var csa models.CSA
			var cbsa models.CBSA
			var _wg sync.WaitGroup
			var mu sync.RWMutex
			workers := make(chan struct{}, CONNECTION_LIMIT)
			results := []serializers.ScoreResults{}
			for i := range scores {
				_wg.Add(1)
				go func(i int) {
					workers <- struct{}{}
					if csa_result := h.DB.Where(&models.CSA{Geoid: scores[i].Geoid}).First(&csa); csa_result.Error != nil {
						csa.CSA_name = ""
					}
					if cbsa_result := h.DB.Where(&models.CBSA{Geoid: scores[i].Geoid}).First(&cbsa); cbsa_result.Error != nil {
						cbsa.CBSA_name = ""
					}
					mu.Lock()
					results = append(
						results,
						serializers.ScoreResults{
							RankID:       scores[i].ID,
							Geoid:        scores[i].Geoid,
							CBSAName:     cbsa.CBSA_name,
							NWI:          scores[i].NWI,
							TransitScore: scores[i].TransitScore,
							BikeScore:    scores[i].BikeScore,
							Format:       query.Format,
						},
					)
					mu.Unlock()
					defer func() {
						_wg.Done()
						<-workers
					}()
				}(i)
			}
			_wg.Wait()

			caches.CACHE.Put(ctx.Request.URL.RequestURI(), results)
			ctx.JSON(http.StatusOK, &results)
			return
		case "xml":
			var csa models.CSA
			var cbsa models.CBSA
			var _wg sync.WaitGroup
			var mu sync.RWMutex
			workers := make(chan struct{}, CONNECTION_LIMIT)
			resultScores := []serializers.ScoreResultsXML{}
			for i := range scores {
				_wg.Add(1)
				go func(i int) {
					if csa_result := h.DB.Where(&models.CSA{Geoid: scores[i].Geoid}).First(&csa); csa_result.Error != nil {
						csa.CSA_name = ""
					}
					if cbsa_result := h.DB.Where(&models.CBSA{Geoid: scores[i].Geoid}).First(&cbsa); cbsa_result.Error != nil {
						cbsa.CBSA_name = ""
					}
					mu.Lock()
					resultScores = append(
						resultScores,
						serializers.ScoreResultsXML{
							XMLName:      xml.Name{Space: "result"},
							RankID:       scores[i].ID,
							Geoid:        scores[i].Geoid,
							CBSAName:     cbsa.CBSA_name,
							NWI:          scores[i].NWI,
							TransitScore: scores[i].TransitScore,
							BikeScore:    scores[i].BikeScore,
							Format:       query.Format,
						},
					)
					mu.Unlock()
					defer func() {
						_wg.Done()
						<-workers
					}()
				}(i)
			}
			_wg.Wait()
			results := serializers.XMLResults{Scores: resultScores}
			caches.CACHE.Put(ctx.Request.URL.RequestURI(), results)
			ctx.XML(http.StatusOK, &results)
			return
		}
	default:
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("bad request! make sure all required fields are filled"))
		return
	}
}

func (h handler) GetScoreDetails(ctx *gin.Context) {
	var query serializers.DetailQuery
	searchedRank := &models.Rank{}
	stringRankID := ctx.Param("id")
	if err := ctx.Bind(&query); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("bad request! make sure you've provided valid parameters: %s", err))
		return
	}
	cbsa := &models.CBSA{}
	if stringRankID == "" {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("bad request! provide rankID as id parameter"))
		return
	}
	rankID, err := strconv.ParseUint(stringRankID, 10, 64)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("bad request! provide a valid rankID"))
		return
	}

	if rankResult := h.DB.Model(&models.Rank{}).Where(&models.Rank{Model: gorm.Model{ID: uint(rankID)}}).First(&searchedRank); rankResult.Error != nil {
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("not found! make sure you have provided a valid rankID: %s", rankResult.Error))
		return
	}
	if cbsaResult := h.DB.Model(&models.CBSA{}).Where(&models.CBSA{Geoid: searchedRank.Geoid}).First(&cbsa); cbsaResult.Error != nil {
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("not found!: %s", cbsaResult.Error))
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
				ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid field name: %s", field))
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
