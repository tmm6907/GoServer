package api

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"reflect"
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

func (h handler) GetScores(ctx *gin.Context) {
	url := ctx.Request.URL.RequestURI()
	cacheResult, ok := caches.CACHE.Get(url)
	if ok {
		switch t := reflect.TypeOf(cacheResult); t {
		case reflect.TypeOf([]serializers.ScoreResults{}), reflect.TypeOf(serializers.AddressScoreResult{}):
			ctx.JSON(http.StatusOK, &cacheResult)
			return
		case reflect.TypeOf(serializers.AddressScoreResultXML{}), reflect.TypeOf(serializers.XMLResults{}):
			ctx.XML(http.StatusOK, &cacheResult)
			return
		default:
			panic(t)
		}
	}
	var query serializers.ScoreQuery
	err := ctx.Bind(&query)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	query.SetFormat()
	switch {
	case query.Address != "":
		var wg sync.WaitGroup
		wg.Add(1)
		geoid := make(chan string)
		go func() {
			wg.Done()
			res, err := query.GetGeoid()
			if err != nil {
				ctx.AbortWithError(http.StatusNotFound, err)
				return
			}
			geoid <- res
		}()
		var score models.Rank
		var cbsa models.CBSA
		geoid10, err := strconv.ParseUint(<-geoid, 10, 64)
		if err != nil {
			ctx.AbortWithError(http.StatusNotFound, err)
			return
		}
		if result := h.DB.Where(&models.Rank{Geoid: geoid10}).First(&score); result.Error != nil {
			ctx.AbortWithError(http.StatusNotFound, result.Error)
			return
		}
		if result := h.DB.Where(&models.CBSA{Geoid: geoid10}).First(&cbsa); result.Error != nil {
			ctx.AbortWithError(http.StatusNotFound, result.Error)
			return
		}
		switch strings.ToLower(query.Format) {
		case "json":
			result := serializers.AddressScoreResult{
				RankID:          score.ID,
				Geoid:           geoid10,
				CBSAName:        cbsa.CBSA_name,
				NWI:             score.NWI,
				TransitScore:    score.TransitScore,
				BikeScore:       score.BikeScore,
				SearchedAddress: query.Address,
				Format:          query.Format,
			}
			caches.CACHE.Put(url, result)
			ctx.JSON(http.StatusOK, &result)
			return
		case "xml":
			result := serializers.AddressScoreResultXML{
				XMLName:         xml.Name{Space: "result"},
				RankID:          score.ID,
				Geoid:           geoid10,
				CBSAName:        cbsa.CBSA_name,
				NWI:             score.NWI,
				TransitScore:    score.TransitScore,
				BikeScore:       score.BikeScore,
				SearchedAddress: query.Address,
				Format:          query.Format,
			}
			caches.CACHE.Put(url, result)
			ctx.XML(http.StatusOK, &result)
			return
		default:
		}

		wg.Wait()
	case query.ZipCode != "":
		var zipScores []models.Rank
		var res []models.ZipCode
		var scores []models.Rank
		query.SetLimit()
		if result := h.DB.Where(&models.ZipCode{Zipcode: query.ZipCode}).Find(&res); result.Error != nil {
			ctx.AbortWithError(http.StatusNotFound, result.Error)
			return
		}
		var subwg sync.WaitGroup
		var mu sync.RWMutex
		for _, item := range res {
			subwg.Add(1)
			go func(item models.ZipCode) {
				defer subwg.Done()
				if result := h.DB.Limit(query.Limit).Offset(query.Offset).Where("cbsas.cbsa = ?", item.CBSA).Model(&models.Rank{}).Select("*").Joins("left join cbsas on cbsas.geoid = ranks.geoid").Scan(&zipScores); result.Error != nil {
					fmt.Println(result.Error)
				}
				mu.Lock()
				scores = append(scores, zipScores...)
				mu.Unlock()
			}(item)
		}
		subwg.Wait()
		switch query.Format {
		case "json":
			var subwg sync.WaitGroup
			var mu sync.RWMutex
			results := []serializers.ScoreResults{}
			for i := range scores {
				subwg.Add(1)
				go func(i int) {
					defer subwg.Done()
					var csa models.CSA
					var cbsa models.CBSA
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
							ID:           i + query.Offset,
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
				}(i)
				subwg.Wait()
			}
			caches.CACHE.Put(url, results)
			ctx.JSON(http.StatusOK, &results)
			return
		case "xml":
			var subwg sync.WaitGroup
			var mu sync.RWMutex
			resultScores := []serializers.ScoreResultsXML{}
			for i := range scores {
				subwg.Add(1)
				go func(i int) {
					defer subwg.Done()
					var csa models.CSA
					var cbsa models.CBSA
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
							ID:           i + query.Offset,
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
				}(i)
				subwg.Wait()
			}
			results := serializers.XMLResults{Scores: resultScores}
			caches.CACHE.Put(url, results)
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
				D2a_ranked:              searchedRank.D2a_ranked,
				D2b_ranked:              searchedRank.D2b_ranked,
				D3b_ranked:              searchedRank.D3b_ranked,
				D4a_ranked:              searchedRank.D4a_ranked,
				PublicTransitEstimate:   cbsa.PublicTransitEstimate,
				PublicTransitPercentage: cbsa.PublicTransitPercentage,
				BikeRidership:           cbsa.BikeRidership,
				BikeRidershipPercentage: cbsa.BikeRidershipPercentage,
			}
			ctx.JSON(http.StatusOK, &result)
			return
		case "xml":
			result := serializers.DetailResultXML{
				RankID:                  searchedRank.ID,
				D2b_e8mixa:              searchedRank.D2b_e8mixa,
				D2a_ephhm:               searchedRank.D2a_ephhm,
				D3b:                     searchedRank.D3b,
				D4a:                     searchedRank.D4a,
				D2a_ranked:              searchedRank.D2a_ranked,
				D2b_ranked:              searchedRank.D2b_ranked,
				D3b_ranked:              searchedRank.D3b_ranked,
				D4a_ranked:              searchedRank.D4a_ranked,
				PublicTransitEstimate:   cbsa.PublicTransitEstimate,
				PublicTransitPercentage: cbsa.PublicTransitPercentage,
				BikeRidership:           cbsa.BikeRidership,
				BikeRidershipPercentage: cbsa.BikeRidershipPercentage,
			}
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
			ctx.JSON(http.StatusOK, &result)
			return
		case "xml":
			result["format"] = "xml"
			result["result"] = xml.Name{}
			ctx.XML(http.StatusOK, &result)
			return
		default:
		}

	}
}

// func (h handler) testEndpoint(ctx *gin.Context) {
// 	param := ctx.Param("id")
// 	id, err := strconv.ParseInt(param, 10, 64)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	var count int64
// 	var ranks []models.Rank
// 	h.DB.Model(&models.Rank{}).Where("geoid = ?", id).Find(&ranks).Count(&count)
// 	fmt.Println("Number of values", count)
// 	ctx.JSON(http.StatusOK, &ranks)
// }
