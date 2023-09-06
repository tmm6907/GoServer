package main

import (
	"log"
	"net/http"
	"os"

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
	CREATE_ZIPCODES
	CREATE_BIKESHARES
	CREATE_FATALITIES
	UPDATE_POPULATION
	UPDATE_TRANSIT_SCORES
	UPDATE_BIKE_SCORES
	UPDATE_BIKE_PERCENT_RANKS
	UPDATE_BIKE_COUNT_RANKS
	UPDATE_BIKESHARE_RANKS
	UPDATE_FATALITIES_RANKS
	CREATE_CBSA_DATAFRAME
	ACTIVATE_RELEASE_MODE
)

func handleFileError(err error, file string) {
	if err != nil {
		log.Fatalf("Error, file %s could not be read", file)
	}
}
func main() {
	flags := ACTIVATE_RELEASE_MODE
	path := "opennwi.db"
	router := gin.Default()
	if flags&ACTIVATE_RELEASE_MODE != 0 {
		gin.SetMode(gin.ReleaseMode)
		path = "/home/nwi/go/go-web/opennwi.db"
	}
	port := os.Getenv("INTERNAL_PORT")
	if port == "" {
		log.Fatalf("Fatal Error in nwi.go: %s environment variable not set.", "INTERNAL_PORT")
	}
	if flags&ACTIVATE_RELEASE_MODE != 0 {
		router.Use(middleware.AuthenticateRequest())
	}
	gormDB, err := db.InitDB(path)
	if err != nil {
		log.Fatalln(err)
	}
	router.Use(middleware.HandleCachedResults())
	if flags&CREATE_DB != 0 {
		dbFile, err := api.ReadData(DB_FILE)
		handleFileError(err, DB_FILE)
		db.RepopulateGroupTracts(gormDB, dbFile)
	}

	if flags&UPDATE_TRANSIT != 0 {
		transit_file, err := api.ReadData(CBSA_TRANSIT_FILE)
		handleFileError(err, CBSA_TRANSIT_FILE)
		db.AddTransitUsage(gormDB, transit_file)
	}

	if flags&UPDATE_BIKE != 0 {
		bike_file, err := api.ReadData(CBSA_BIKE_FILE)
		handleFileError(err, CBSA_BIKE_FILE)
		db.AddBikeRidership(gormDB, bike_file)
	}

	if flags&UPDATE_BIKE_PERCENT != 0 {
		db.FindBikeRidershipPercentage(gormDB)
	}

	if flags&UPDATE_POPULATION != 0 {
		popFile, err := api.ReadData(POPULATION_FILE)
		handleFileError(err, POPULATION_FILE)
		db.AddCBSAPopulation(gormDB, popFile)
	}

	if flags&CREATE_ZIPCODES != 0 {
		zip_file, err := api.ReadData(ZIPCODE_FILE)
		handleFileError(err, ZIPCODE_FILE)
		db.CreateZips(gormDB, zip_file)
	}

	if flags&CREATE_BIKESHARES != 0 {
		db.CreateBikeShares(gormDB)
	}

	if flags&CREATE_FATALITIES != 0 {
		db.CreateFatalities(gormDB)
	}

	if flags&UPDATE_BIKE_COUNT_RANKS != 0 {
		db.AddBikeCountRanks(gormDB)
	}

	if flags&UPDATE_BIKESHARE_RANKS != 0 {
		db.AddBikeShareRanks(gormDB)
	}

	if flags&UPDATE_BIKE_PERCENT_RANKS != 0 {
		db.AddBikePercentageRanks(gormDB)
	}

	if flags&UPDATE_FATALITIES_RANKS != 0 {
		db.AddFatalityRanks(gormDB)
	}

	if flags&UPDATE_BIKE_SCORES != 0 {
		db.AddBikeScores(gormDB)
	}

	if flags&UPDATE_TRANSIT_SCORES != 0 {
		db.AddTransitScores(gormDB)
	}

	if flags&CREATE_CBSA_DATAFRAME != 0 {
		db.WriteToCBSADataframe(gormDB)
	}
	api.RegisterRoutes(router, gormDB)
	router.GET("/", func(c *gin.Context) {
		endpoints := []string{"/scores/", "/details/"}
		c.JSON(http.StatusOK, gin.H{
			"message":   "Welcome to OpenNWI, provider of local and regional walk, bike and tranist score!",
			"endpoints": endpoints,
		})
	})
	router.Run(port)
}
