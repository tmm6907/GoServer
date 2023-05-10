package group_tracts

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"
)

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

func CreateTractGroups(database [][]string) []GroupTract {
	var census_tract_groups []GroupTract
	for _, record := range database {
		record[30] = record[30][16 : len(record[30])-3]
		geoid10, err := strconv.ParseUint(record[1], 10, 64)
		if err != nil {
			geoid10 = 0
		}

		geoid20, err := strconv.ParseUint(record[2], 10, 64)
		if err != nil {
			geoid20 = 0
		}
		statefp, err := strconv.ParseUint(record[3], 10, 8)
		if err != nil {
			statefp = 0
		}
		countryfp, err := strconv.ParseUint(record[4], 10, 16)
		if err != nil {
			countryfp = 0
		}
		tractce, err := strconv.ParseUint(record[5], 10, 32)
		if err != nil {
			tractce = 0
		}
		blkgrpce, err := strconv.ParseUint(record[6], 10, 8)
		if err != nil {
			blkgrpce = 0
		}
		csa, err := strconv.ParseUint(record[7], 10, 16)
		if err != nil {
			csa = 0
		}
		cbsa, err := strconv.ParseUint(record[9], 10, 32)
		if err != nil {
			cbsa = 0
		}
		ac_t, err := strconv.ParseFloat(record[11], 64)
		if err != nil {
			ac_t = 0
		}
		ac_w, err := strconv.ParseFloat(record[12], 64)
		if err != nil {
			ac_w = 0
		}
		ac_l, err := strconv.ParseFloat(record[13], 64)
		if err != nil {
			ac_l = 0
		}
		ac_u, err := strconv.ParseFloat(record[14], 64)
		if err != nil {
			ac_u = 0
		}
		totp, err := strconv.ParseUint(record[15], 10, 16)
		if err != nil {
			totp = 0
		}
		counthu, err := strconv.ParseFloat(record[16], 64)
		if err != nil {
			counthu = 0
		}
		hh, err := strconv.ParseFloat(record[17], 64)
		if err != nil {
			hh = 0
		}
		workers, err := strconv.ParseUint(record[18], 10, 16)
		if err != nil {
			workers = 0
		}
		d2b, err := strconv.ParseFloat(record[19], 64)
		if err != nil {
			d2b = 0
		}
		d2a, err := strconv.ParseFloat(record[20], 64)
		if err != nil {
			d2a = 0
		}
		d3b, err := strconv.ParseFloat(record[21], 64)
		if err != nil {
			d3b = 0
		}
		d4a, err := strconv.ParseFloat(record[22], 64)
		if err != nil {
			d4a = 0
		}
		d2a_r, err := strconv.ParseFloat(record[23], 64)
		if err != nil {
			d2a_r = 0
		}
		d2b_r, err := strconv.ParseFloat(record[24], 64)
		if err != nil {
			d2b_r = 0
		}
		d3b_r, err := strconv.ParseFloat(record[25], 64)
		if err != nil {
			d3b_r = 0
		}
		d4a_r, err := strconv.ParseFloat(record[26], 64)
		if err != nil {
			d4a_r = 0
		}
		nwi, err := strconv.ParseFloat(record[27], 64)
		if err != nil {
			nwi = 0
		}
		sh_l, err := strconv.ParseFloat(record[28], 64)
		if err != nil {
			sh_l = 0
		}
		sh_a, err := strconv.ParseFloat(record[29], 64)
		if err != nil {
			sh_a = 0
		}
		group_tract := GroupTract{
			Geoid10: geoid10,
			Geoid20: geoid20,
			GeoidDetail: GeoidDetail{
				Statefp:   uint8(statefp),
				Countryfp: uint16(countryfp),
				Tractce:   uint32(tractce),
				Blkgrpce:  uint8(blkgrpce)},
			CSA: CSA{
				CSA:      uint16(csa),
				CSA_name: record[8],
			},
			CBSA: CBSA{
				CBSA:      uint32(cbsa),
				CBSA_name: record[10],
			},
			AC: AC{
				AC_total: ac_t,
				AC_water: ac_w,
				AC_land:  ac_l,
				AC_unpr:  ac_u,
			},
			Population: Population{
				Total_pop: uint16(totp),
				CountHU:   counthu,
				HH:        hh,
				Workers:   uint16(workers),
			},
			Rank: Rank{
				D2b_e8mixa: d2b,
				D2a_ephhm:  d2a,
				D3b:        d3b,
				D4a:        d4a,
				D2a_ranked: float32(d2a_r),
				D2b_ranked: float32(d2b_r),
				D3b_ranked: float32(d3b_r),
				D4a_ranked: float32(d4a_r),
				NWI:        nwi,
			},
			Shape: Shape{
				Shape_length: sh_l,
				Shape_area:   sh_a,
				Geometry:     record[30],
			},
		}
		census_tract_groups = append(census_tract_groups, group_tract)
	}
	return census_tract_groups
}

func MatchZipToCBSA(file string) []Zipcode {
	var zipcodes []Zipcode
	records, err := ReadData(file)
	if err != nil {
		log.Fatalln(err)
	}
	for _, record := range records {
		cbsa, err := strconv.ParseUint(record[7], 10, 32)
		if err != nil {
			cbsa = 0
		}
		zipcodes = append(zipcodes, Zipcode{
			Zipcode: record[0],
			CBSA:    uint32(cbsa),
		})
	}
	return zipcodes
}
