package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"nwi.io/nwi/api"
	"nwi.io/nwi/models"
)

const ZIPCODE_FILE = "Zip_To_CBSA.csv"
const DB_FILE = "NWI.csv"
const CBSA_TRANSIT_FILE = "CBSA_Public_Transit_Usage.csv"
const CBSA_BIKE_FILE = "CBSA_Bicylce_Ridership.csv"
const INSERT_LIMIT = 100

type DBEntry interface {
	models.BlockGroup | models.ZipCode
}

func create_entry[N DBEntry](db *gorm.DB, data []N) *gorm.DB {
	for i := 0; i < len(data); i += INSERT_LIMIT {
		result := db.Create(data[i : i+INSERT_LIMIT])
		if result.Error != nil {
			fmt.Println(result.Error)
		}
		fmt.Print(result.RowsAffected)
	}
	return db
}

func addTransitUsage(db *gorm.DB, database [][]string, wg *sync.WaitGroup) {
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
		db.Model(&models.CBSA{}).Where("CBSA = ?", uint32(cbsa_id)).Updates(models.CBSA{PublicTansitEstimate: uint64(usageEstimate)})
	}

}

func addTransitScores(db *gorm.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	var cbsas []models.CBSA
	var quantiles api.Quantile
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
		go func(i int) {
			rank := api.GetTransitScore(cbsas[i].PublicTansitPercentage, quantiles)
			db.Model(&models.Rank{}).Where(&models.Rank{Geoid: cbsas[i].Geoid}).Updates(models.Rank{TransitScore: uint8(rank)})
		}(i)
	}
}

func addBikeRidership(db *gorm.DB, database [][]string, wg *sync.WaitGroup) {
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
		db.Model(&models.CBSA{}).Where("CBSA = ?", uint32(cbsa_id)).Updates(models.CBSA{BikeRidership: usage})
	}

}
func createZipToCBSA(db *gorm.DB, database [][]string, wg *sync.WaitGroup) {
	defer wg.Done()
	zipcodes := api.MatchZipToCBSA(database)
	result := create_entry(db, zipcodes)
	if result.Error != nil {
		log.Println(result.Error)
	}
}

func repopulateGroupTracts(db *gorm.DB, database [][]string, wg *sync.WaitGroup) {
	defer wg.Done()
	db_data := make(chan []models.BlockGroup)
	go func() {
		res := api.CreateTractGroups(database)
		db_data <- res
	}()
	res := <-db_data
	result := create_entry(db, res)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
}
func init_db(url string) (*gorm.DB, error) {
	// Initialize
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

func main() {
	gin.SetMode(gin.ReleaseMode)
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		log.Fatalf("Fatal Error in connect_unix.go: %s environment variable not set.", dbUser)
	}
	dbPass := os.Getenv("DB_PASS")
	if dbPass == "" {
		log.Fatalf("Fatal Error in connect_unix.go: %s environment variable not set.", dbPass)
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		log.Fatalf("Fatal Error in connect_unix.go: %s environment variable not set.", dbName)
	}
	connectionName := os.Getenv("CLOUD_SQL_CONNECTION_NAME")
	if connectionName == "" {
		log.Fatalf("Fatal Error in connect_unix.go: %s environment variable not set.", connectionName)
	}
	port := os.Getenv("INTERNAL_PORT")
	if port == "" {
		log.Fatalf("Fatal Error in connect_unix.go: %s environment variable not set.", port)
	}
	// "/cloudsql/"+connectionName,
	dbUrl := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?parseTime=true",
		dbUser,
		dbPass,
		"localhost",
		dbName,
	)
	db, err := init_db(dbUrl)
	if err != nil {
		log.Fatalln(err)
	}
	// var wg sync.WaitGroup
	// wg.Add(1)

	// db_file, err := api.ReadData(DB_FILE)
	// if err != nil {
	// 	log.Fatalf("Error, file %s could not be read", db_file)
	// }
	// go repopulateGroupTracts(db, db_file, &wg)

	// transit_file, err := api.ReadData(CBSA_TRANSIT_FILE)
	// if err != nil {
	// 	log.Fatalf("Error, file %s could not be read", transit_file)
	// }
	// go addTransitUsage(db, transit_file, &wg)

	// bike_file, err := api.ReadData(CBSA_BIKE_FILE)
	// if err != nil {
	// 	log.Fatalf("Error, file %s could not be read", bike_file)
	// }
	// go addBikeRidership(db, bike_file, &wg)

	// zip_file, err := api.ReadData(ZIPCODE_FILE)
	// if err != nil {
	// 	log.Fatalf("Error, file %s could not be read", zip_file)
	// }
	// go createZipToCBSA(db, zip_file, &wg)

	// go addTransitScores(db, &wg)
	router := gin.Default()
	api.RegisterRoutes(router, db)
	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"title":     "Open-NWI API",
			"body":      "Welcome to Open-NWI API, An open-source API to access EPA's National Walkability Index for any address recognized by US Census Geocoding.",
			"endpoints": "/scores",
			"examples": gin.H{
				"default":         "https://opennwi.dev/scores/",
				"fiteredScores":   "https://opennwi.dev/scores?limit=3&offset=1200",
				"searchByAddress": "https://opennwi.dev/scores?address=1600%20Pennsylvania%20Avenue%20Northwest%20Washington%20DC",
				"searchByZipcode": "https://opennwi.dev/scores?zipcode=20024",
			},
		},
		)
		// wg.Wait()
	})
	router.Run(port)
}
