package db

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"nwi.io/nwi/api"
	"nwi.io/nwi/models"
)

const INSERT_LIMIT = 500
const CONNECTION_LIMIT = 1000

type DBEntry interface {
	models.BlockGroup | models.ZipCode | models.State | models.BikeShare | models.BikeFatalities
}

func CreateDBEntry[N DBEntry](db *gorm.DB, data []N) *gorm.DB {
	for i := 0; i < len(data); i += INSERT_LIMIT {
		if len(data) > INSERT_LIMIT {
			result := db.Create(data[i : i+INSERT_LIMIT])
			if result.Error != nil {
				log.Fatalln(result.Error)
			}
		} else {
			result := db.Create(data)
			if result.Error != nil {
				log.Fatalln(result.Error)
			}
		}
	}
	return db
}

func AddTransitUsage(db *gorm.DB, database [][]string) {
	for _, record := range database {
		cbsa_id, err := strconv.ParseUint(record[4], 10, 32)
		if err != nil {
			log.Fatalln(cbsa_id, " not found")
		}
		usageEstimate, err := strconv.ParseFloat(record[3], 64)
		if err != nil {
			usageEstimate = 0
		}
		usagePercentage, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			usagePercentage = 0
		}
		db.Model(&models.CBSA{}).Where(&models.CBSA{CBSA: uint32(cbsa_id)}).Updates(models.CBSA{
			PublicTransitPercentage: usagePercentage,
			PublicTransitEstimate:   uint64(usageEstimate),
		})
	}
}

func AddTransitScores(db *gorm.DB) {
	var cbsas []models.CBSA
	var quantiles api.Quantile
	cbsaResult := db.Where("public_transit_percentage > 0").Find(&cbsas)
	if cbsaResult.Error != nil {
		fmt.Print(cbsaResult.Error)
		return
	}
	quantiles_data, err := api.ReadData("Transit_Ranks.csv")
	if err != nil {
		fmt.Print(err)
	}
	for i := range quantiles_data {
		if i != 0 {
			q, err := strconv.ParseFloat(quantiles_data[i][1], 64)
			if err != nil {
				log.Fatalln(err)
			}
			quantiles = append(quantiles, q)
		}
	}
	for i := range cbsas {
		searchRank := api.GetScores(cbsas[i].PublicTransitPercentage, quantiles)
		if rank_result := db.Model(&models.Rank{}).Where(&models.Rank{Geoid: cbsas[i].Geoid}).Updates(models.Rank{TransitScore: uint8(searchRank)}); rank_result.Error != nil {
			log.Fatalln(rank_result.Error)
		}
	}
}

func CreateBikeShares(db *gorm.DB) {
	var shares []models.BikeShare
	var subwg sync.WaitGroup
	bikeShareData, err := api.ReadData("BikeShareData.csv")
	if err != nil {
		panic(err)
	}
	for i, record := range bikeShareData {
		if i != 0 {
			fips, _ := strconv.ParseUint(record[0], 10, 32)
			count, _ := strconv.ParseUint(record[2], 10, 16)
			share := models.BikeShare{
				FIPS:  uint32(fips),
				Count: uint16(count),
			}
			shares = append(shares, share)
		}
	}
	CreateDBEntry(db, shares)
}

func AddBikeShareRanks(db *gorm.DB) {
	var bikeShares []models.BikeShare
	var quantiles api.Quantile
	bikeShareResult := db.Where("count > 0").Find(&bikeShares)
	if bikeShareResult.Error != nil {
		fmt.Print(bikeShareResult.Error)
		return
	}
	quantiles_data, err := api.ReadData("BikeShare_Ranks.csv")
	if err != nil {
		fmt.Print(err)
	}
	for i := range quantiles_data {
		if i != 0 {
			q, err := strconv.ParseFloat(quantiles_data[i][1], 64)
			if err != nil {
				log.Fatalln(err)
			}
			quantiles = append(quantiles, q)
		}
	}
	for _, bikeShare := range bikeShares {
		searchRank := api.GetScores(uint(bikeShare.Count), quantiles)
		if rank_result := db.Model(&models.Rank{}).Where(fmt.Sprintf("geoid LIKE '%v%%'", bikeShare.FIPS)).Updates(models.Rank{BikeShareRank: uint8(searchRank)}); rank_result.Error != nil {
			log.Fatalln(rank_result.Error)
		}
	}
}

func AddBikePercentageRanks(db *gorm.DB) {
	var cbsas []models.CBSA
	var quantiles api.Quantile
	cbsaResult := db.Where("bike_ridership_percentage > 0").Find(&cbsas)
	if cbsaResult.Error != nil {
		fmt.Print(cbsaResult.Error)
		return
	}
	quantiles_data, err := api.ReadData("Bike_Ridership_Ranks.csv")
	if err != nil {
		fmt.Print(err)
	}

	for i := range quantiles_data {
		if i != 0 {
			q, err := strconv.ParseFloat(quantiles_data[i][1], 64)
			if err != nil {
				log.Fatalln(err)
			}
			quantiles = append(quantiles, q)
		}
	}
	for i := range cbsas {
		searchRank := api.GetScores(cbsas[i].BikeRidershipPercentage, quantiles)
		if rank_result := db.Where(&models.Rank{Geoid: cbsas[i].Geoid}).Updates(models.Rank{BikePercentageRank: uint8(searchRank)}); rank_result.Error != nil {
			log.Fatalln(rank_result.Error)
		}
	}
}

func AddBikeScores(db *gorm.DB) {
	var ranks []models.Rank
	totalRankResult := db.Find(&ranks)
	if totalRankResult.Error != nil {
		fmt.Print(totalRankResult.Error)
		return
	}
	for _, rank := range ranks {
		searchRank := api.CalculateBikeScore(rank.BikeCountRank, rank.BikePercentageRank, rank.BikeFatalityRank, rank.BikeShareRank)
		if rank_result := db.Model(&models.Rank{}).Where(&models.Rank{Geoid: rank.Geoid}).Updates(models.Rank{BikeScore: searchRank}); rank_result.Error != nil {
			fmt.Println(rank_result.Error)
		}
	}
}

func CreateFatalities(db *gorm.DB) {
	var states []models.State
	fatalityData, err := api.ReadData("BikeFatalities.csv")
	if err != nil {
		fmt.Print(err)
	}
	for i, record := range fatalityData {
		if i != 0 {
			percentage, _ := strconv.ParseFloat(record[4], 32)
			statefp, _ := strconv.ParseUint(record[6], 10, 8)
			state := models.State{
				Name:    record[0],
				Statefp: uint8(statefp),
				BikeFatalities: models.BikeFatalities{
					FatalityPercentage: float32(percentage),
				},
			}
			states = append(states, state)
		}
	}
	CreateDBEntry(db, states)
}

func AddFatalityRanks(db *gorm.DB) {
	var fatalities []models.BikeFatalities
	var quantiles api.Quantile
	fatalityResult := db.Where("fatality_percentage > 0").Find(&fatalities)
	if fatalityResult.Error != nil {
		fmt.Print(fatalityResult.Error)
		return
	}
	quantiles_data, err := api.ReadData("BikeFatality_Ranks.csv")
	if err != nil {
		fmt.Print(err)
	}
	for i := range quantiles_data {
		if i != 0 {
			q, err := strconv.ParseFloat(quantiles_data[i][1], 64)
			if err != nil {
				log.Fatalln(err)
			}
			quantiles = append(quantiles, q)
		}
	}
	var geoidDetail []models.GeoidDetail
	for _, fatality := range fatalities {
		searchRank := api.GetScores(float64(fatality.FatalityPercentage), quantiles)
		if cbsaResult := db.Where(&models.GeoidDetail{Statefp: fatality.Statefp}).Find(&geoidDetail); cbsaResult.Error != nil {
			log.Fatalln(cbsaResult.Error)
		}
		for _, item := range geoidDetail {
			if rank_result := db.Model(&models.Rank{}).Where(&models.Rank{Geoid: item.Geoid}).Updates(models.Rank{BikeFatalityRank: uint8(searchRank)}); rank_result.Error != nil {
				log.Fatalln(rank_result.Error)
			}
		}
	}
}

func AddBikeCountRanks(db *gorm.DB) {
	var cbsas []models.CBSA
	var quantiles api.Quantile
	cbsaResult := db.Where("bike_ridership > 0").Find(&cbsas)
	if cbsaResult.Error != nil {
		fmt.Print(cbsaResult.Error)
		return
	}
	quantiles_data, err := api.ReadData("Bike_Ranks.csv")
	if err != nil {
		fmt.Print(err)
	}
	for i := range quantiles_data {
		if i != 0 {
			q, err := strconv.ParseFloat(quantiles_data[i][1], 64)
			if err != nil {
				log.Fatalln(err)
			}
			quantiles = append(quantiles, q)
		}
	}
	for _, cbsa := range cbsas {
		searchRank := api.GetScores(uint(cbsa.BikeRidership), quantiles)
		if rank_result := db.Model(&models.Rank{}).Where(&models.Rank{Geoid: cbsa.Geoid}).Updates(models.Rank{BikeCountRank: uint8(searchRank)}); rank_result.Error != nil {
			fmt.Println(rank_result.Error)
		}
	}
}

func AddBikeRidership(db *gorm.DB, database [][]string) {
	for _, record := range database {
		cbsa_id, err := strconv.ParseUint(record[3], 10, 64)
		if err != nil {
			cbsa_id = 99999
		}

		usage, err := strconv.ParseUint(record[2], 10, 64)
		if err != nil {
			usage = 0
		}

		if cbsa_result := db.Where(&models.CBSA{CBSA: uint32(cbsa_id)}).Updates(&models.CBSA{BikeRidership: usage}); cbsa_result.Error != nil {
			time.Sleep(2 * time.Second)
			db.Where(&models.CBSA{CBSA: uint32(cbsa_id)}).Updates(&models.CBSA{BikeRidership: usage})
		}
	}
	
}

func AddCBSAPopulation(db *gorm.DB, database [][]string) {
	for _, record := range database {
		cbsa_id, err := strconv.ParseUint(record[0], 10, 64)
		if err != nil {
			cbsa_id = 99999
		}

		pop, err := strconv.ParseUint(record[8], 10, 64)
		if err != nil {
			pop = 0
		}

		if cbsaResult := db.Model(&models.CBSA{}).Where("CBSA = ? and population = 0", uint32(cbsa_id)).Updates(&models.CBSA{Population: pop}); cbsaResult.Error != nil {
			log.Fatalln(cbsaResult.Error)
		}
	}
}

func FindBikeRidershipPercentage(db *gorm.DB) {
	
	var cbsas []models.CBSA
	cbsaResult := db.Where("population > 0 and bike_ridership_percentage = 0 and bike_ridership > 0").Find(&cbsas)
	if cbsaResult.Error != nil {
		fmt.Print(cbsaResult.Error)
		return
	}

	for _, item := range cbsas {
		precentageResult := db.Model(&item).Update("bike_ridership_percentage", float64(item.BikeRidership)/float64(item.Population))
		if precentageResult.Error != nil{
			fmt.Print(precentageResult.Error)
			return
		}
	}
}

func CreateZips(db *gorm.DB, database [][]string) {
	db_data := make(chan []models.ZipCode)
	go func() {
		res := api.MatchZipToCBSA(database)
		db_data <- res
	}()
	res := <-db_data
	result := CreateDBEntry(db, res)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
}

func RepopulateGroupTracts(db *gorm.DB, database [][]string) {
	db_data := make(chan []models.BlockGroup)
	go func() {
		res := api.CreateTractGroups(database)
		db_data <- res
	}()
	res := <-db_data
	result := CreateDBEntry(db, res)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
}

func WriteToCBSADataframe(db *gorm.DB) {
	var cbsas []models.CBSA
	db.Find(&cbsas)
	file, err := os.Create("Bike_Ridership_Ranks.csv")
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"CBSA", "Ridership_Percentage"}
	writer.Write(header)
	for _, item := range cbsas {
		row := []string{fmt.Sprint(item.CBSA), strconv.FormatFloat(item.BikeRidershipPercentage, 'f', -1, 64)}
		writer.Write(row)
	}
}

func InitDB(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	result := db.Exec("PRAGMA journal_mode = WAL")
    if result.Error != nil {
		return nil, result.Error
    }

	db.AutoMigrate(
		&models.BlockGroup{},
		&models.GeoidDetail{},
		&models.CSA{},
		&models.CBSA{},
		&models.AC{},
		&models.Population{},
		&models.Rank{},
		&models.Shape{},
		&models.ZipCode{},
		&models.State{},
		&models.BikeFatalities{},
		&models.BikeShare{},
	)
	return db, nil
}
