package main

import (
	"fmt"
	"log"
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

func handleFileError(err error, file string) {
	if err != nil {
		log.Fatalf("Error, file %s could not be read", file)
	}
}
func main() {
	gin.SetMode(gin.ReleaseMode)
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

	// "/cloudsql/"+connectionName,
	dbUrl := fmt.Sprintf(
		"%s:%s@unix(%s)/%s?parseTime=true",
		dbUser,
		dbPass,
		"/cloudsql/"+connectionName,
		dbName,
	)

	// dbUrl := fmt.Sprintf(
	// 	"postgresql://%s:%s@%s?sslmode=verify-full",
	// 	cockroachDbUser,
	// 	cockroachDbPass,
	// 	cockroachDbConnection,
	// )

	gormDB, err := db.InitDB(dbUrl)
	if err != nil {
		log.Fatalln(err)
	}
	// var wg sync.WaitGroup
	// wg.Add(1)

	// dbFile, err := api.ReadData(DB_FILE)
	// handleFileError(err, DB_FILE)
	// go db.RepopulateGroupTracts(gormDB, dbFile, &wg)

	// transit_file, err := api.ReadData(CBSA_TRANSIT_FILE)
	// handleFileError(err, CBSA_TRANSIT_FILE)
	// go db.AddTransitUsage(gormDB, transit_file, &wg)

	// bike_file, err := api.ReadData(CBSA_BIKE_FILE)
	// handleFileError(err, CBSA_BIKE_FILE)
	// go db.AddBikeRidership(gormDB, bike_file, &wg)

	// zip_file, err := api.ReadData(ZIPCODE_FILE)
	// handleFileError(err, ZIPCODE_FILE)
	// go db.CreateZipToCBSA(gormDB, zip_file, &wg)

	// go db.AddTransitScores(gormDB, &wg)

	// popFile, err := api.ReadData(POPULATION_FILE)
	// handleFileError(err, POPULATION_FILE)
	// go db.AddCBSAPopulation(gormDB, popFile, &wg)

	router := gin.Default()
	router.Use(middleware.AuthenticateRequest())
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
		// wg.Wait()
	})
	router.Run(port)
}
