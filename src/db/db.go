package db

import (
	"fmt"
	"log"
	"strconv"
	"sync"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"nwi.io/nwi/api"
	"nwi.io/nwi/models"
)

const INSERT_LIMIT = 100

type DBEntry interface {
	models.BlockGroup | models.ZipCode
}

func CreateDBEntry[N DBEntry](db *gorm.DB, data []N) *gorm.DB {
	for i := 0; i < len(data); i += INSERT_LIMIT {
		result := db.Create(data[i : i+INSERT_LIMIT])
		if result.Error != nil {
			fmt.Println(result.Error)
		}
		fmt.Print(result.RowsAffected)
	}
	return db
}

func AddTransitUsage(db *gorm.DB, database [][]string, wg *sync.WaitGroup) {
	defer wg.Done()

	for _, record := range database {
		cbsa_id, err := strconv.ParseUint(record[4], 10, 32)
		if err != nil {
			cbsa_id = 99999
		}
		usageEstimate, err := strconv.ParseFloat(record[3], 64)
		if err != nil {
			fmt.Println(err)
			usageEstimate = 0
		}
		usagePercentage, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			fmt.Println(err)
			usagePercentage = 0
		}
		db.Model(&models.CBSA{}).Where(&models.CBSA{CBSA: uint32(cbsa_id)}).Updates(models.CBSA{
			PublicTansitPercentage: usagePercentage,
			PublicTansitEstimate:   uint64(usageEstimate),
		})
	}

}

func AddTransitScores(db *gorm.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	var cbsas []models.CBSA
	var quantiles api.Quantile
	var rank models.Rank
	cbsaResult := db.Find(&cbsas)
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
				fmt.Println(err)
			}
			quantiles = append(quantiles, q)
		}
	}
	for i := range cbsas {
		searchRank := api.GetTransitScore(cbsas[i].PublicTansitPercentage, quantiles)
		if rank_result := db.Where(&models.Rank{Geoid: cbsas[i].Geoid}).First(&rank); rank_result.Error != nil {
			fmt.Println(rank_result.Error)
		}
		rank.TransitScore = uint8(searchRank)
		db.Save(&rank)
	}
}

func AddBikeRidership(db *gorm.DB, database [][]string, wg *sync.WaitGroup) {
	defer wg.Done()
	for _, record := range database {
		cbsa_id, err := strconv.ParseUint(record[3], 10, 64)
		if err != nil {
			fmt.Println(err)
			cbsa_id = 99999
		}

		usage, err := strconv.ParseUint(record[2], 10, 64)
		if err != nil {
			fmt.Println(err)
			usage = 0
		}
		if cbsa_result := db.Where(&models.CBSA{CBSA: uint32(cbsa_id)}).Updates(&models.CBSA{BikeRidership: usage}); cbsa_result.Error != nil {
			fmt.Println(cbsa_result.Error)
		}
	}

}

func AddCBSAPopulation(db *gorm.DB, database [][]string, wg *sync.WaitGroup) {
	defer wg.Done()
	for _, record := range database {
		cbsa_id, err := strconv.ParseUint(record[0], 10, 64)
		if err != nil {
			fmt.Println(err)
			cbsa_id = 99999
		}

		pop, err := strconv.ParseUint(record[8], 10, 64)
		if err != nil {
			fmt.Println(err)
			pop = 0
		}
		if cbsa_result := db.Model(&models.CBSA{}).Where(&models.CBSA{CBSA: uint32(cbsa_id)}).Updates(&models.CBSA{Population: pop}); cbsa_result.Error != nil {
			fmt.Println(cbsa_result.Error)
		}
	}

}

func RepopulateGroupTracts(db *gorm.DB, database [][]string, wg *sync.WaitGroup) {
	defer wg.Done()
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
func InitDB(url string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(url))
	if err != nil {
		return nil, err
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
	)
	return db, nil
}
