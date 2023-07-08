package api

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"nwi.io/nwi/caches"
	"nwi.io/nwi/models"
	"nwi.io/nwi/serializers"
)

type XMLResults struct {
	XMLName xml.Name                     `xml:"results"`
	Scores  []serializers.ScoreResultXML `xml:"scores"`
}

func getGeoid(address string) (string, error) {
	var geoidResults serializers.GeoCodingResult
	url := "https://geocoding.geo.census.gov/geocoder/geographies/onelineaddress?address=" + address + "&benchmark=2020&vintage=Census2010_Census2020&format=json"
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Golang_Spider_Bot/3.0")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(body, &geoidResults); err != nil { // Parse []byte to go struct pointer
		return "", err
	}
	if len(geoidResults.Result.AddressMatches) > 0 {
		return geoidResults.Result.AddressMatches[0].Geographies.CensusBlocks[0].Geoid[:len(geoidResults.Result.AddressMatches[0].Geographies.CensusBlocks[0].Geoid)-3], nil
	} else {
		return "", nil
	}
}

func (h handler) GetScores(ctx *gin.Context) {
	url := ctx.Request.URL.RequestURI()
	cacheResult, ok := caches.CACHE.Get(url)
	if ok {
		switch t := reflect.TypeOf(cacheResult); t {
		case reflect.TypeOf([]serializers.ScoreResult{}), reflect.TypeOf(serializers.AddressScoreResult{}):
			ctx.JSON(http.StatusOK, &cacheResult)
			return
		case reflect.TypeOf(serializers.AddressScoreResultXML{}), reflect.TypeOf(XMLResults{}):
			ctx.XML(http.StatusOK, &cacheResult)
			return
		default:
			fmt.Println(t)
			fmt.Println(reflect.TypeOf(XMLResults{}))
			fmt.Println(ctx.Request.URL.Path)
		}
	}
	address := strings.ReplaceAll(ctx.Query("address"), " ", "%20")
	zipcode := ctx.Query("zipcode")
	format := ctx.Query("format")
	if zipcode != "" && address != "" {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
	if zipcode == "" && address == "" {
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
	switch {
	case address != "":
		var wg sync.WaitGroup
		wg.Add(1)
		geoid := make(chan string)
		go func() {
			wg.Done()
			res, err := getGeoid(address)
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
		switch format {
		case "json":
			result := serializers.AddressScoreResult{
				Geoid:                          geoid10,
				NWI:                            score.NWI,
				SearchedAddress:                ctx.Query("address"),
				RegionalTransitUsagePercentage: cbsa.PublicTansitPercentage,
				RegionalTransitUsage:           cbsa.PublicTansitEstimate,
				RegionalBikeRidership:          cbsa.BikeRidership,
				Format:                         format,
			}
			caches.CACHE.Put(url, result)
			ctx.JSON(http.StatusOK, &result)
		case "xml":
			result := serializers.AddressScoreResultXML{
				XMLName:                        xml.Name{Space: "result"},
				Geoid:                          geoid10,
				NWI:                            score.NWI,
				SearchedAddress:                ctx.Query("address"),
				RegionalTransitUsagePercentage: cbsa.PublicTansitPercentage,
				RegionalTransitUsage:           cbsa.PublicTansitEstimate,
				RegionalBikeRidership:          cbsa.BikeRidership,
				Format:                         format,
			}
			caches.CACHE.Put(url, result)
			ctx.XML(http.StatusOK, &result)
		default:
			ctx.AbortWithStatus(http.StatusBadRequest)
		}

		wg.Wait()
	case zipcode != "":
		var zipScores []models.Rank
		var res []models.ZipCode
		var scores []models.Rank
		if len(zipcode) > 5 {
			ctx.AbortWithStatus(http.StatusBadRequest)
		}
		limit, limit_err := strconv.Atoi(ctx.Query("limit"))
		if limit_err != nil {
			limit = 50
		}
		offset, offset_err := strconv.Atoi(ctx.Query("offset"))
		if offset_err != nil {
			offset = 0
		}
		if zipcode == "" {
			if result := h.DB.Limit(limit).Offset(offset).Find(&scores); result.Error != nil {
				ctx.AbortWithError(http.StatusNotFound, result.Error)
				return
			}
		} else {
			if result := h.DB.Where(&models.ZipCode{Zipcode: zipcode}).Find(&res); result.Error != nil {
				ctx.AbortWithError(http.StatusNotFound, result.Error)
				return
			}
			var subwg sync.WaitGroup
			var mu sync.RWMutex
			for _, item := range res {
				subwg.Add(1)
				go func(item models.ZipCode) {
					defer subwg.Done()
					if result := h.DB.Limit(limit).Offset(offset).Where("cbsas.cbsa = ?", item.CBSA).Model(&models.Rank{}).Select("*").Joins("left join cbsas on cbsas.geoid = ranks.geoid").Scan(&zipScores); result.Error != nil {
						fmt.Println(result.Error)
					}
					mu.Lock()
					scores = append(scores, zipScores...)
					mu.Unlock()
				}(item)
			}
			subwg.Wait()
		}
		switch format {
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
							ID:                             i + offset,
							Geoid:                          scores[i].Geoid,
							CSA_name:                       csa.CSA_name,
							CBSA_name:                      cbsa.CBSA_name,
							NWI:                            scores[i].NWI,
							RegionalTransitUsagePercentage: cbsa.PublicTansitPercentage,
							RegionalTransitUsage:           cbsa.PublicTansitEstimate,
							RegionalBikeRidership:          cbsa.BikeRidership,
							Format:                         format,
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
							ID:                             i + offset,
							Geoid:                          scores[i].Geoid,
							CSA_name:                       csa.CSA_name,
							CBSA_name:                      cbsa.CBSA_name,
							NWI:                            scores[i].NWI,
							RegionalTransitUsagePercentage: cbsa.PublicTansitPercentage,
							RegionalTransitUsage:           cbsa.PublicTansitEstimate,
							RegionalBikeRidership:          cbsa.BikeRidership,
							Format:                         format,
						},
					)
					mu.Unlock()
				}(i)
				subwg.Wait()
			}
			results := XMLResults{Scores: resultScores}
			caches.CACHE.Put(url, results)
			ctx.XML(http.StatusOK, &results)
		}
	default:
		ctx.AbortWithStatus(http.StatusBadRequest)
	}
}
