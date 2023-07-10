package api

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"nwi.io/nwi/caches"
	"nwi.io/nwi/models"
	"nwi.io/nwi/serializers"
)

func (h handler) GetScores(ctx *gin.Context) {
	url := ctx.Request.URL.RequestURI()
	cacheResult, ok := caches.CACHE.Get(url)
	if ok {
		switch t := reflect.TypeOf(cacheResult); t {
		case reflect.TypeOf([]serializers.ScoreResult{}), reflect.TypeOf(serializers.AddressScoreResult{}):
			ctx.JSON(http.StatusOK, &cacheResult)
			return
		case reflect.TypeOf(serializers.AddressScoreResultXML{}), reflect.TypeOf(serializers.XMLResults{}):
			ctx.XML(http.StatusOK, &cacheResult)
			return
		default:
			fmt.Println(t)
			fmt.Println(reflect.TypeOf(serializers.XMLResults{}))
			fmt.Println(ctx.Request.URL.Path)
		}
	}
	var query serializers.ScoreQuery
	err := ctx.Bind(&query)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
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
		switch query.Format {
		case "json":
			result := serializers.AddressScoreResult{
				Geoid:                          geoid10,
				NWI:                            score.NWI,
				SearchedAddress:                query.Address,
				RegionalTransitUsagePercentage: cbsa.PublicTansitPercentage,
				RegionalTransitUsage:           cbsa.PublicTansitEstimate,
				RegionalBikeRidership:          cbsa.BikeRidership,
				Format:                         query.Format,
			}
			caches.CACHE.Put(url, result)
			ctx.JSON(http.StatusOK, &result)
		case "xml":
			result := serializers.AddressScoreResultXML{
				XMLName:                        xml.Name{Space: "result"},
				Geoid:                          geoid10,
				NWI:                            score.NWI,
				SearchedAddress:                query.Address,
				RegionalTransitUsagePercentage: cbsa.PublicTansitPercentage,
				RegionalTransitUsage:           cbsa.PublicTansitEstimate,
				RegionalBikeRidership:          cbsa.BikeRidership,
				Format:                         query.Format,
			}
			caches.CACHE.Put(url, result)
			ctx.XML(http.StatusOK, &result)
		default:
			ctx.AbortWithStatus(http.StatusBadRequest)
		}

		wg.Wait()
	case query.ZipCode != "":
		var zipScores []models.Rank
		var res []models.ZipCode
		var scores []models.Rank
		if result := h.DB.Where(&models.ZipCode{Zipcode: query.ZipCode}).Find(&res); result.Error != nil {
			ctx.AbortWithError(http.StatusNotFound, result.Error)
			return
		}
		fmt.Println(len(res))
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
			results := []serializers.ScoreResult{}
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
						serializers.ScoreResult{
							ID:                             i + query.Offset,
							Geoid:                          scores[i].Geoid,
							CSA_name:                       csa.CSA_name,
							CBSA_name:                      cbsa.CBSA_name,
							NWI:                            scores[i].NWI,
							RegionalTransitUsagePercentage: cbsa.PublicTansitPercentage,
							RegionalTransitUsage:           cbsa.PublicTansitEstimate,
							RegionalBikeRidership:          cbsa.BikeRidership,
							Format:                         query.Format,
						},
					)
					mu.Unlock()
				}(i)
				subwg.Wait()
			}
			caches.CACHE.Put(url, results)
			ctx.JSON(http.StatusOK, &results)
		case "xml":
			var subwg sync.WaitGroup
			var mu sync.RWMutex
			resultScores := []serializers.ScoreResultXML{}
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
						serializers.ScoreResultXML{
							XMLName:                        xml.Name{Space: "result"},
							ID:                             i + query.Offset,
							Geoid:                          scores[i].Geoid,
							CSA_name:                       csa.CSA_name,
							CBSA_name:                      cbsa.CBSA_name,
							NWI:                            scores[i].NWI,
							RegionalTransitUsagePercentage: cbsa.PublicTansitPercentage,
							RegionalTransitUsage:           cbsa.PublicTansitEstimate,
							RegionalBikeRidership:          cbsa.BikeRidership,
							Format:                         query.Format,
						},
					)
					mu.Unlock()
				}(i)
				subwg.Wait()
			}
			results := serializers.XMLResults{Scores: resultScores}
			caches.CACHE.Put(url, results)
			ctx.XML(http.StatusOK, &results)
		}
	default:
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
}
