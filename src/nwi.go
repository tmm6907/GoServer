package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"nwi.io/nwi/api"
	"nwi.io/nwi/db"
	"nwi.io/nwi/middleware"
)

const ZIPCODE_FILE = "Zip_To_CBSA.csv"
const DB_FILE = "NWI.csv"
const CBSA_TRANSIT_FILE = "CBSA_Public_Transit_Usage.csv"
const CBSA_BIKE_FILE = "CBSA_Bicylce_Ridership.csv"
const POPULATION_FILE = "2022_CBSA_POP_ESTIMATE.csv"

const (
	CREATE_DB = 1 << iota
	UPDATE_TRANSIT
	UPDATE_BIKE
	UPDATE_BIKE_PERCENT
	CREATE_ZIPCODE
	UPDATE_POPULATION
	UPDATE_TRANSIT_SCORES
	UPDATE_BIKE_SCORES
	CREATE_CBSA_DATAFRAME
	ACTIVATE_RELEASE_MODE
)

func handleFileError(err error, file string) {
	if err != nil {
		log.Fatalf("Error, file %s could not be read", file)
	}
}
func main() {
	var wg sync.WaitGroup
	router := gin.Default()
	flags := ACTIVATE_RELEASE_MODE

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		log.Fatalf("Fatal Error in nwi.go: %s environment variable not set.", "DB_USER")
	}
	dbPass := os.Getenv("DB_PASS")
	if dbPass == "" {
		log.Fatalf("Fatal Error in nwi.go: %s environment variable not set.", "DB_PASS")
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		log.Fatalf("Fatal Error in nwi.go: %s environment variable not set.", "DB_NAME")
	}
	connectionName := os.Getenv("CLOUD_SQL_CONNECTION_NAME")
	if connectionName == "" {
		log.Fatalf("Fatal Error in nwi.go: %s environment variable not set.", "CLOUD_SQL_CONNECTION_NAME")
	}
	port := os.Getenv("INTERNAL_PORT")
	if port == "" {
		log.Fatalf("Fatal Error in nwi.go: %s environment variable not set.", "INTERNAL_PORT")
	}
	dbUrl := ""
	if flags&ACTIVATE_RELEASE_MODE != 0 {
		gin.SetMode(gin.ReleaseMode)
		dbUrl = fmt.Sprintf(
			"%s:%s@unix(%s)/%s?parseTime=true",
			dbUser,
			dbPass,
			"/cloudsql/"+connectionName,
			dbName,
		)
		router.Use(middleware.AuthenticateRequest())
	} else {
		dbUrl = fmt.Sprintf(
			"%s:%s@tcp(%s)/%s?parseTime=true",
			dbUser,
			dbPass,
			"localhost",
			dbName,
		)
	}

	gormDB, err := db.InitDB(dbUrl)
	if err != nil {
		log.Fatalln(err)
	}

	if flags&CREATE_DB != 0 {
		wg.Add(1)
		dbFile, err := api.ReadData(DB_FILE)
		handleFileError(err, DB_FILE)
		go db.RepopulateGroupTracts(gormDB, dbFile, &wg)
	}

	if flags&UPDATE_TRANSIT != 0 {
		wg.Add(1)
		transit_file, err := api.ReadData(CBSA_TRANSIT_FILE)
		handleFileError(err, CBSA_TRANSIT_FILE)
		go db.AddTransitUsage(gormDB, transit_file, &wg)
	}

	if flags&UPDATE_BIKE != 0 {
		wg.Add(1)
		bike_file, err := api.ReadData(CBSA_BIKE_FILE)
		handleFileError(err, CBSA_BIKE_FILE)
		go db.AddBikeRidership(gormDB, bike_file, &wg)
	}

	if flags&UPDATE_BIKE_PERCENT != 0 {
		wg.Add(1)
		go db.FindBikeRidershipPercentage(gormDB, &wg)
	}

	if flags&UPDATE_POPULATION != 0 {
		wg.Add(1)
		popFile, err := api.ReadData(POPULATION_FILE)
		handleFileError(err, POPULATION_FILE)
		go db.AddCBSAPopulation(gormDB, popFile, &wg)
	}

	if flags&CREATE_ZIPCODE != 0 {
		wg.Add(1)
		zip_file, err := api.ReadData(ZIPCODE_FILE)
		handleFileError(err, ZIPCODE_FILE)
		go db.CreateZips(gormDB, zip_file, &wg)
	}

	if flags&UPDATE_TRANSIT_SCORES != 0 {
		wg.Add(1)
		go db.AddTransitScores(gormDB, &wg)
	}

	if flags&UPDATE_BIKE_SCORES != 0 {
		wg.Add(1)
		go db.AddBikeScores(gormDB, &wg)
	}

	if flags&CREATE_CBSA_DATAFRAME != 0 {
		wg.Add(1)
		go db.WriteToCBSADataframe(gormDB, &wg)
	}

	api.RegisterRoutes(router, gormDB)
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
	})
	router.Run(port)
	wg.Wait()
}
