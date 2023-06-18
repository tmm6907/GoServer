package api

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"nwi.io/nwi/models"
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

func GetTransitScore(percentage float64, quantile Quantile) int {
	for i := range quantile {
		rank := i + 1
		if percentage < quantile[i] {
			return rank - 1
		}
	}
	return 0
}

func CreateTractGroups(database [][]string) []models.BlockGroup {
	var census_tract_groups []models.BlockGroup
	for _, record := range database {
		// record[30] = record[30][16 : len(record[30])-3]

		_geoid10, err := strconv.ParseUint(record[geoid10], 10, 64)
		if err != nil {
			_geoid10 = 0
		}
		_geoid20, err := strconv.ParseUint(record[geoid20], 10, 64)
		if err != nil {
			_geoid20 = 0
		}

		_statefp, err := strconv.ParseUint(record[statefp], 10, 8)
		if err != nil {
			_statefp = 0
		}
		_countyfp, err := strconv.ParseUint(record[countyfp], 10, 16)
		if err != nil {
			_countyfp = 0
		}
		_tractce, err := strconv.ParseUint(record[tractce], 10, 32)
		if err != nil {
			_tractce = 0
		}
		_blkgrpce, err := strconv.ParseUint(record[blkgrpce], 10, 8)
		if err != nil {
			_blkgrpce = 0
		}
		_csa, err := strconv.ParseFloat(record[csa], 64)
		if err != nil {
			fmt.Println(err)
			_csa = 0
		}
		_cbsa, err := strconv.ParseFloat(record[cbsa], 32)
		if err != nil {
			fmt.Println(err)
			_cbsa = 0
		}
		ac_t, err := strconv.ParseFloat(record[acTotal], 64)
		if err != nil {
			ac_t = 0
		}
		ac_w, err := strconv.ParseFloat(record[acWater], 64)
		if err != nil {
			ac_w = 0
		}
		ac_l, err := strconv.ParseFloat(record[acLand], 64)
		if err != nil {
			ac_l = 0
		}
		ac_u, err := strconv.ParseFloat(record[acUnpr], 64)
		if err != nil {
			ac_u = 0
		}
		totp, err := strconv.ParseUint(record[totalPop], 10, 16)
		if err != nil {
			totp = 0
		}
		counthu, err := strconv.ParseFloat(record[countHU], 64)
		if err != nil {
			counthu = 0
		}
		_hh, err := strconv.ParseFloat(record[hh], 64)
		if err != nil {
			_hh = 0
		}
		_cbsaPop, err := strconv.ParseUint(record[cbsaPop], 10, 16)
		if err != nil {
			_cbsaPop = 0
		}
		_d2b, err := strconv.ParseFloat(record[d2b], 64)
		if err != nil {
			_d2b = 0
		}
		_d2a, err := strconv.ParseFloat(record[d2a], 64)
		if err != nil {
			_d2a = 0
		}
		_d3b, err := strconv.ParseFloat(record[d3b], 64)
		if err != nil {
			_d3b = 0
		}
		_d4a, err := strconv.ParseFloat(record[d4a], 64)
		if err != nil {
			_d4a = 0
		}
		d2a_r, err := strconv.ParseFloat(record[d2aRanked], 64)
		if err != nil {
			d2a_r = 0
		}
		d2b_r, err := strconv.ParseFloat(record[d2bRanked], 64)
		if err != nil {
			d2b_r = 0
		}
		d3b_r, err := strconv.ParseFloat(record[d3bRanked], 64)
		if err != nil {
			d3b_r = 0
		}
		d4a_r, err := strconv.ParseFloat(record[d4aRanked], 64)
		if err != nil {
			d4a_r = 0
		}
		_nwi, err := strconv.ParseFloat(record[nwi], 64)
		if err != nil {
			_nwi = 0
		}
		sh_l, err := strconv.ParseFloat(record[shapeLength], 64)
		if err != nil {
			sh_l = 0
		}
		sh_a, err := strconv.ParseFloat(record[shapeArea], 64)
		if err != nil {
			sh_a = 0
		}
		group_tract := models.BlockGroup{
			Geoid10: _geoid10,
			Geoid20: _geoid20,
			GeoidDetail: models.GeoidDetail{
				Statefp:  uint8(_statefp),
				Countyfp: uint16(_countyfp),
				Tractce:  uint32(_tractce),
				Blkgrpce: uint8(_blkgrpce)},
			CSA: models.CSA{
				CSA:      uint16(_csa),
				CSA_name: record[csaName],
			},
			CBSA: models.CBSA{
				CBSA:       uint32(_cbsa),
				CBSA_name:  record[cbsaName],
				Population: _cbsaPop,
			},
			AC: models.AC{
				AC_total: ac_t,
				AC_water: ac_w,
				AC_land:  ac_l,
				AC_unpr:  ac_u,
			},
			Population: models.Population{
				Total_pop: uint16(totp),
				CountHU:   counthu,
				HH:        _hh,
			},
			Rank: models.Rank{
				D2b_e8mixa: _d2b,
				D2a_ephhm:  _d2a,
				D3b:        _d3b,
				D4a:        _d4a,
				D2a_ranked: float32(d2a_r),
				D2b_ranked: float32(d2b_r),
				D3b_ranked: float32(d3b_r),
				D4a_ranked: float32(d4a_r),
				NWI:        _nwi,
			},
			Shape: models.Shape{
				Shape_length: sh_l,
				Shape_area:   sh_a,
			},
		}
		fmt.Println(group_tract.CBSA.CBSA)
		census_tract_groups = append(census_tract_groups, group_tract)
	}
	return census_tract_groups
}
