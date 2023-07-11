package serializers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type ScoreQuery struct {
	Address string `form:"address" json:"address" xml:"address" binding:"required_without=ZipCode"`
	ZipCode string `form:"zipcode" json:"zipcode" xml:"zipcode" validate:"min=5,max=5"`
	Limit   int    `form:"limit" json:"limit" xml:"limit" validate:"max=500"`
	Offset  int    `form:"offset" json:"offset" xml:"offset"`
	Format  string `form:"format" json:"format" xml:"format" binding:"required"`
}

func (q *ScoreQuery) GetGeoid() (string, error) {
	var geoidResults GeoCodingResult
	url := "https://geocoding.geo.census.gov/geocoder/geographies/onelineaddress?address=" + q.GetAddress() + "&benchmark=2020&vintage=Census2010_Census2020&format=json"
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
func (q *ScoreQuery) GetAddress() string {
	return strings.ReplaceAll(q.Address, " ", "%20")
}
func (q *ScoreQuery) SetLimit() {
	if q.Limit == 0 {
		q.Limit = 50
	}
}

func (q *ScoreQuery) SetFormat() {
	if q.Format == "" {
		q.Format = "json"
	}
}
