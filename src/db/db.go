package db

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"nwi.io/nwi/api"
	"nwi.io/nwi/models"
)

const INSERT_LIMIT = 500
const CONNECTION_LIMIT = 2000

type DBEntry interface {
	models.BlockGroup | models.ZipCode
}

func CreateDBEntry[N DBEntry](db *gorm.DB, data []N) *gorm.DB {
	var wg sync.WaitGroup
	workers := make(chan struct{}, CONNECTION_LIMIT)
	for i := 0; i < len(data); i += INSERT_LIMIT {
		wg.Add(1)
		go func(i int) {
			workers <- struct{}{}
			result := db.Create(data[i : i+INSERT_LIMIT])
			if result.Error != nil {
				fmt.Println(result.Error)
			}
			defer func() {
				<-workers
				wg.Done()
			}()
		}(i)
	}
	wg.Wait()
	return db
}

func AddTransitUsage(db *gorm.DB, database [][]string, wg *sync.WaitGroup) {
	defer wg.Done()

	var subwg sync.WaitGroup
	workers := make(chan struct{}, CONNECTION_LIMIT)
	for _, record := range database {
		cbsa_id, err := strconv.ParseUint(record[4], 10, 32)
		if err != nil {
			fmt.Println(cbsa_id, " not found")
		}
		usageEstimate, err := strconv.ParseFloat(record[3], 64)
		if err != nil {
			usageEstimate = 0
		}
		usagePercentage, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			usagePercentage = 0
		}
		subwg.Add(1)
		go func(cbsa_id uint64, usageEstimate float64, usagePercentage float64) {
			workers <- struct{}{}

			db.Model(&models.CBSA{}).Where(&models.CBSA{CBSA: uint32(cbsa_id)}).Updates(models.CBSA{
				PublicTransitPercentage: usagePercentage,
				PublicTransitEstimate:   uint64(usageEstimate),
			})
			defer func() {
				<-workers
				subwg.Done()
			}()
		}(cbsa_id, usageEstimate, usagePercentage)
	}
	subwg.Wait()
}

func AddTransitScores(db *gorm.DB, wg *sync.WaitGroup) {
	defer wg.Done()
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
				fmt.Println(err)
			}
			quantiles = append(quantiles, q)
		}
	}
	var subwg sync.WaitGroup
	workers := make(chan struct{}, CONNECTION_LIMIT)
	for i := range cbsas {
		subwg.Add(1)
		go func(i int) {
			workers <- struct{}{}
			searchRank := api.GetScores(cbsas[i].PublicTransitPercentage, quantiles)
			if rank_result := db.Model(&models.Rank{}).Where(&models.Rank{Geoid: cbsas[i].Geoid}).Updates(models.Rank{TransitScore: uint8(searchRank)}); rank_result.Error != nil {
				fmt.Println(rank_result.Error)
			}
			defer func() {
				<-workers
				subwg.Done()
			}()
		}(i)
	}
	subwg.Wait()
}

func AddBikeScores(db *gorm.DB, wg *sync.WaitGroup) {
	defer wg.Done()
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
				fmt.Println(err)
			}
			quantiles = append(quantiles, q)
		}
	}
	var subwg sync.WaitGroup
	workers := make(chan struct{}, CONNECTION_LIMIT)
	for i := range cbsas {
		subwg.Add(1)
		go func(i int) {
			workers <- struct{}{}
			searchRank := api.GetScores(cbsas[i].BikeRidershipPercentage, quantiles)
			if rank_result := db.Where(&models.Rank{Geoid: cbsas[i].Geoid}).Updates(models.Rank{BikeScore: uint8(searchRank)}); rank_result.Error != nil {
				fmt.Println(rank_result.Error)
			}
			defer func() {
				<-workers
				subwg.Done()
			}()
		}(i)
	}
	subwg.Wait()
}

func AddBikeRidership(db *gorm.DB, database [][]string, wg *sync.WaitGroup) {
	defer wg.Done()
	var subwg sync.WaitGroup
	workers := make(chan struct{}, CONNECTION_LIMIT)
	for _, record := range database {
		cbsa_id, err := strconv.ParseUint(record[3], 10, 64)
		if err != nil {
			cbsa_id = 99999
		}

		usage, err := strconv.ParseUint(record[2], 10, 64)
		if err != nil {
			usage = 0
		}
		subwg.Add(1)
		go func(record []string, cbsa_id uint64, usage uint64) {
			workers <- struct{}{}
			if cbsa_result := db.Where(&models.CBSA{CBSA: uint32(cbsa_id)}).Updates(&models.CBSA{BikeRidership: usage}); cbsa_result.Error != nil {
				time.Sleep(2 * time.Second)
				db.Where(&models.CBSA{CBSA: uint32(cbsa_id)}).Updates(&models.CBSA{BikeRidership: usage})
			}
			defer func() {
				<-workers
				subwg.Add(1)
			}()
		}(record, cbsa_id, usage)
	}
	subwg.Wait()
}

func AddCBSAPopulation(db *gorm.DB, database [][]string, wg *sync.WaitGroup) {
	defer wg.Done()
	var subwg sync.WaitGroup
	workers := make(chan struct{}, CONNECTION_LIMIT)
	for _, record := range database {
		subwg.Add(1)
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
		go func(record []string, cbsa_id uint64, pop uint64) {
			workers <- struct{}{}

			if cbsa_result := db.Model(&models.CBSA{}).Where("CBSA = ? and population = 0", uint32(cbsa_id)).Updates(&models.CBSA{Population: pop}); cbsa_result.Error != nil {
				fmt.Println(cbsa_result.Error)
			}
			defer func() {
				<-workers
				subwg.Done()
			}()
		}(record, cbsa_id, pop)
	}
	subwg.Wait()
}
func FindBikeRidershipPercentage(db *gorm.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	var cbsas []models.CBSA
	cbsaResult := db.Where("population > 0 and bike_ridership_percentage = 0 and bike_ridership > 0").Find(&cbsas)
	if cbsaResult.Error != nil {
		fmt.Print(cbsaResult.Error)
		return
	}
	var subwg sync.WaitGroup
	workers := make(chan struct{}, CONNECTION_LIMIT)
	for _, item := range cbsas {
		subwg.Add(1)
		go func(item models.CBSA) {
			workers <- struct{}{}

			db.Model(&item).Update("bike_ridership_percentage", float64(item.BikeRidership)/float64(item.Population))
			defer func() {
				<-workers
				subwg.Done()
			}()
		}(item)
	}
	subwg.Wait()
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

func CreateZips(db *gorm.DB, database [][]string, wg *sync.WaitGroup) {
	defer wg.Done()
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

func WriteToCBSADataframe(db *gorm.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	var cbsas []models.CBSA
	db.Find(&cbsas)
	file, err := os.Create("Bike_Ridership_Ranks.csv")
	if err != nil {
		fmt.Println(err)
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
