package api

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"nwi.io/nwi/models"
	"nwi.io/nwi/serializers"
)

const (
	geoid10 = iota + 1
	geoid20
	statefp
	countyfp
	tractce
	blkgrpce
	csa
	csaName
	cbsa
	cbsaName
	cbsaPop
	acTotal
	acWater
	acLand
	acUnpr
	totalPop
	countHU
	hh
	d2b
	d2a
	d3b
	d4a
	d2aRanked
	d2bRanked
	d3bRanked
	d4aRanked
	nwi
	shapeLength
	shapeArea
)

const (
	countFactor     float64 = 0.55
	percentFactor   float64 = 0.35
	bikeShareFactor float64 = 0.10
	fatalityFactor  float64 = 0.15
)

type ScoreInput interface {
	float64 | uint
}
type Quantile []float64

func ReadData(fileName string) ([][]string, error) {

	f, err := os.Open(fileName)

	if err != nil {
		return [][]string{}, err
	}

	defer f.Close()

	r := csv.NewReader(f)
	// skip first line
	if _, err := r.Read(); err != nil {
		return [][]string{}, err
	}

	records, err := r.ReadAll()

	if err != nil {
		return [][]string{}, err
	}

	return records, nil
}

func parseTractGroupFields(record []string) *serializers.TractGroup {
	group := &serializers.TractGroup{}
	intVal, _ := strconv.ParseInt(record[geoid10], 10, 64)
	group.Geoid10 = int(intVal)
	intVal, _ = strconv.ParseInt(record[geoid20], 10, 64)
	group.Geoid20 = int(intVal)
	intVal, _ = strconv.ParseInt(record[statefp], 10, 8)
	group.Statefp = int(intVal)
	intVal, _ = strconv.ParseInt(record[countyfp], 10, 16)
	group.Countyfp = int(intVal)
	intVal, _ = strconv.ParseInt(record[tractce], 10, 32)
	group.Tractce = int(intVal)
	intVal, _ = strconv.ParseInt(record[blkgrpce], 10, 8)
	group.Blkgrpce = int(intVal)
	floatVal, _ := strconv.ParseFloat(record[csa], 64)
	group.CSA = int(floatVal)
	floatVal, _ = strconv.ParseFloat(record[cbsa], 32)
	group.CBSA = int(floatVal)
	floatVal, _ = strconv.ParseFloat(record[acTotal], 64)
	group.ACT = floatVal
	floatVal, _ = strconv.ParseFloat(record[acWater], 64)
	group.ACW = floatVal
	floatVal, _ = strconv.ParseFloat(record[acLand], 64)
	group.ACL = floatVal
	floatVal, _ = strconv.ParseFloat(record[acUnpr], 64)
	group.ACU = floatVal
	intVal, _ = strconv.ParseInt(record[totalPop], 10, 16)
	group.TOTP = int(intVal)
	floatVal, _ = strconv.ParseFloat(record[countHU], 64)
	group.Counthu = floatVal
	floatVal, _ = strconv.ParseFloat(record[hh], 64)
	group.HH = floatVal
	intVal, _ = strconv.ParseInt(record[cbsaPop], 10, 16)
	group.CbsaPop = int(intVal)
	floatVal, _ = strconv.ParseFloat(record[d2b], 64)
	group.D2B = floatVal
	floatVal, _ = strconv.ParseFloat(record[d2a], 64)
	group.D2A = floatVal
	floatVal, _ = strconv.ParseFloat(record[d3b], 64)
	group.D3B = floatVal
	floatVal, _ = strconv.ParseFloat(record[d4a], 64)
	group.D4A = floatVal
	floatVal, _ = strconv.ParseFloat(record[d2aRanked], 64)
	group.D2AR = floatVal
	floatVal, _ = strconv.ParseFloat(record[d2bRanked], 64)
	group.D2BR = floatVal
	floatVal, _ = strconv.ParseFloat(record[d3bRanked], 64)
	group.D3BR = floatVal
	floatVal, _ = strconv.ParseFloat(record[d4aRanked], 64)
	group.D4AR = floatVal
	floatVal, _ = strconv.ParseFloat(record[nwi], 64)
	group.NWI = floatVal
	floatVal, _ = strconv.ParseFloat(record[shapeLength], 64)
	group.SHL = floatVal
	floatVal, _ = strconv.ParseFloat(record[shapeArea], 64)
	group.SHA = floatVal
	return group
}

func CreateTractGroups(database [][]string) []models.BlockGroup {
	var census_tract_groups []models.BlockGroup
	for _, record := range database {
		group := parseTractGroupFields(record)
		group_tract := models.BlockGroup{
			Geoid10: uint64(group.Geoid10),
			Geoid20: uint64(group.Geoid20),
			GeoidDetail: models.GeoidDetail{
				Statefp:  uint8(group.Statefp),
				Countyfp: uint16(group.Countyfp),
				Tractce:  uint32(group.Tractce),
				Blkgrpce: uint8(group.Blkgrpce)},
			CSA: models.CSA{
				CSA:      uint16(group.CSA),
				CSA_name: record[csaName],
			},
			CBSA: models.CBSA{
				CBSA:      uint32(group.CBSA),
				CBSA_name: record[cbsaName],
			},
			AC: models.AC{
				AC_total: group.ACT,
				AC_water: group.ACW,
				AC_land:  group.ACL,
				AC_unpr:  group.ACU,
			},
			Population: models.Population{
				Total_pop: uint16(group.TOTP),
				CountHU:   group.Counthu,
				HH:        group.HH,
			},
			Rank: models.Rank{
				D2b_e8mixa: group.D2B,
				D2a_ephhm:  group.D2A,
				D3b:        group.D3B,
				D4a:        group.D4A,
				D2a_ranked: float32(group.D2AR),
				D2b_ranked: float32(group.D2BR),
				D3b_ranked: float32(group.D3BR),
				D4a_ranked: float32(group.D4AR),
				NWI:        group.NWI,
			},
			Shape: models.Shape{
				Shape_length: group.SHL,
				Shape_area:   group.SHA,
			},
		}
		census_tract_groups = append(census_tract_groups, group_tract)
	}
	return census_tract_groups
}

func MatchZipToCBSA(records [][]string) []models.ZipCode {
	var zipcodes []models.ZipCode
	for _, record := range records {
		cbsa_float, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			cbsa_float = 0
		}
		cbsa := uint32(cbsa_float)
		if len(record[1]) == 4 {
			record[1] = fmt.Sprintf("%s%s", "0", record[1])
		}
		zipcodes = append(zipcodes, models.ZipCode{
			Zipcode: record[1],
			CBSA:    cbsa,
		})
	}
	return zipcodes
}

func GetScores[T ScoreInput](input T, quantile Quantile) int {
	for i := range quantile {
		if float64(input) <= quantile[i] {
			return i + 1
		}
	}
	return 0
}

func CalculateBikeScore(count uint8, percent uint8, fatality uint8, bikeShare uint8) float64 {
	score := (countFactor * float64(count)) + (percentFactor * float64(percent)) + (fatalityFactor * -float64(fatality)) + (bikeShareFactor * float64(bikeShare))
	if score > 0 {
		return score
	}
	if count != 0 || percent != 0 || fatality != 0 || bikeShare != 0 {
		return 1
	}
	return 0
}
