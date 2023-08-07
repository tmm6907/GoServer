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
const CONNECTION_LIMIT = 1000

type DBEntry interface {
	models.BlockGroup | models.ZipCode | models.State | models.BikeShare | models.BikeFatalities
}

func CreateDBEntry[N DBEntry](db *gorm.DB, data []N) *gorm.DB {
	var wg sync.WaitGroup
	workers := make(chan struct{}, CONNECTION_LIMIT)
	for i := 0; i < len(data); i += INSERT_LIMIT {
		wg.Add(1)
		go func(i int) {
			workers <- struct{}{}
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
				log.Fatalln(err)
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
				log.Fatalln(rank_result.Error)
			}
			defer func() {
				<-workers
				subwg.Done()
			}()
		}(i)
	}
	subwg.Wait()
}

func CreateBikeShares(db *gorm.DB, wg *sync.WaitGroup) {
	var shares []models.BikeShare
	defer wg.Done()
	var subwg sync.WaitGroup
	bikeShareData, err := api.ReadData("BikeShareData.csv")
	if err != nil {
		panic(err)
	}
	for i, record := range bikeShareData {
		subwg.Add(1)
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
	subwg.Add(1)
	go CreateDBEntry(db, shares)
	subwg.Wait()
}

func AddBikeShareRanks(db *gorm.DB, wg *sync.WaitGroup) {
	defer wg.Done()
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
	var subwg sync.WaitGroup
	workers := make(chan struct{}, CONNECTION_LIMIT)
	for _, bikeShare := range bikeShares {
		subwg.Add(1)
		go func(bikeShare models.BikeShare) {
			workers <- struct{}{}
			searchRank := api.GetScores(uint(bikeShare.Count), quantiles)
			if rank_result := db.Model(&models.Rank{}).Where(fmt.Sprintf("geoid LIKE '%v%%'", bikeShare.FIPS)).Updates(models.Rank{BikeShareRank: uint8(searchRank)}); rank_result.Error != nil {
				log.Fatalln(rank_result.Error)
			}
			defer func() {
				<-workers
				subwg.Done()
			}()
		}(bikeShare)
	}
	subwg.Wait()
}

func AddBikePercentageRanks(db *gorm.DB, wg *sync.WaitGroup) {
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
				log.Fatalln(err)
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
			if rank_result := db.Where(&models.Rank{Geoid: cbsas[i].Geoid}).Updates(models.Rank{BikePercentageRank: uint8(searchRank)}); rank_result.Error != nil {
				log.Fatalln(rank_result.Error)
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
	var ranks []models.Rank
	totalRankResult := db.Find(&ranks)
	if totalRankResult.Error != nil {
		fmt.Print(totalRankResult.Error)
		return
	}
	var subwg sync.WaitGroup
	workers := make(chan struct{}, CONNECTION_LIMIT)
	for _, rank := range ranks {
		subwg.Add(1)
		go func(rank models.Rank) {
			workers <- struct{}{}
			searchRank := api.CalculateBikeScore(rank.BikeCountRank, rank.BikePercentageRank, rank.BikeFatalityRank, rank.BikeShareRank)
			if rank_result := db.Model(&models.Rank{}).Where(&models.Rank{Geoid: rank.Geoid}).Updates(models.Rank{BikeScore: searchRank}); rank_result.Error != nil {
				fmt.Println(rank_result.Error)
			}
			defer func() {
				<-workers
				subwg.Done()
			}()
		}(rank)
	}
	subwg.Wait()
}

func CreateFatalities(db *gorm.DB, wg *sync.WaitGroup) {
	var states []models.State
	var createwg sync.WaitGroup
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
	createwg.Add(1)
	go CreateDBEntry(db, states)
	createwg.Wait()
}

func AddFatalityRanks(db *gorm.DB, wg *sync.WaitGroup) {
	defer wg.Done()
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
	var subwg sync.WaitGroup
	var geoidDetail []models.GeoidDetail
	workers := make(chan struct{}, CONNECTION_LIMIT)
	for _, fatality := range fatalities {
		subwg.Add(1)
		go func(fatality models.BikeFatalities) {
			workers <- struct{}{}
			searchRank := api.GetScores(float64(fatality.FatalityPercentage), quantiles)
			if cbsaResult := db.Where(&models.GeoidDetail{Statefp: fatality.Statefp}).Find(&geoidDetail); cbsaResult.Error != nil {
				log.Fatalln(cbsaResult.Error)
			}
			for _, item := range geoidDetail {
				if rank_result := db.Model(&models.Rank{}).Where(&models.Rank{Geoid: item.Geoid}).Updates(models.Rank{BikeFatalityRank: uint8(searchRank)}); rank_result.Error != nil {
					log.Fatalln(rank_result.Error)
				}
			}
			defer func() {
				<-workers
				subwg.Done()
			}()
		}(fatality)
	}
	subwg.Wait()
}

func AddBikeCountRanks(db *gorm.DB, wg *sync.WaitGroup) {
	defer wg.Done()
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
	var subwg sync.WaitGroup
	workers := make(chan struct{}, CONNECTION_LIMIT)
	for _, cbsa := range cbsas {
		subwg.Add(1)
		go func(cbsa models.CBSA) {
			workers <- struct{}{}
			searchRank := api.GetScores(uint(cbsa.BikeRidership), quantiles)
			if rank_result := db.Model(&models.Rank{}).Where(&models.Rank{Geoid: cbsa.Geoid}).Updates(models.Rank{BikeCountRank: uint8(searchRank)}); rank_result.Error != nil {
				fmt.Println(rank_result.Error)
			}
			defer func() {
				<-workers
				subwg.Done()
			}()
		}(cbsa)
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
			cbsa_id = 99999
		}

		pop, err := strconv.ParseUint(record[8], 10, 64)
		if err != nil {
			pop = 0
		}
		go func(record []string, cbsa_id uint64, pop uint64) {
			workers <- struct{}{}

			if cbsa_result := db.Model(&models.CBSA{}).Where("CBSA = ? and population = 0", uint32(cbsa_id)).Updates(&models.CBSA{Population: pop}); cbsa_result.Error != nil {
				log.Fatalln(cbsa_result.Error)
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

// func setUpSSL(dsn string, mysqlClientKey string, mysqlCaCert string, mysqlClientCert string) {
// 	var isTLS bool
// 	if mysqlClientKey != "" && mysqlCaCert != "" && mysqlClientCert != "" {
// 		isTLS = true
// 		rootCertPool := x509.NewCertPool()
// 		pem, err := os.ReadFile("/home/tmm6907/.ssl/root/example.crt")
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
// 			log.Fatal("Failed to append PEM.")
// 		}
// 		clientCert := make([]tls.Certificate, 0, 1)
// 		certs, err := tls.LoadX509KeyPair("/path/mysqlClientCert", "/path/mysqlClientKey")
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		clientCert = append(clientCert, certs)

// 		mysql.RegisterTLSConfig("custom", &tls.Config{
// 			RootCAs:      rootCertPool,
// 			Certificates: clientCert,
// 		})
// 	}

// 	// try to connect to mysql database.
// 	cfg := mysql.Config{}

// 	if isTLS == true {
// 		cfg.TLSConfig = "custom"
// 	}

// 	str := cfg.FormatDSN()

//		db, err := gorm.Open("mysql", str)
//	}
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
		&models.State{},
		&models.BikeFatalities{},
		&models.BikeShare{},
	)
	return db, nil
}
