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

func create_entry[T []models.BlockGroup | []models.ZipCode](db *gorm.DB, data T) *gorm.DB {
	result := db.CreateInBatches(data, 50)
	if result.Error != nil {
		log.Fatalln(result.Error)
	}
	return result
}

func addTransitUsage(db *gorm.DB, database [][]string, wg *sync.WaitGroup) {
	defer wg.Done()
	var cbsas []models.CBSA

	for _, record := range database {
		id, err := strconv.ParseUint(record[4], 10, 64)
		if err != nil {
			id = 99999
		}
		result := db.Where(&models.CBSA{CBSA: uint32(id)}).Find(&cbsas)
		if result.Error != nil {
			log.Fatalln(result.Error)
		}
		usage, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			usage = 0
		}
		for _, cbsa := range cbsas {
			if usage != 0 && cbsa.PublicTansitUsage == 0 {
				cbsa.PublicTansitUsage = usage
			}
			db.Save(&cbsa)
		}
	}

}

func addBikeRidership(db *gorm.DB, database [][]string, wg *sync.WaitGroup) {
	defer wg.Done()
	var cbsas []models.CBSA
	for _, record := range database {
		id, err := strconv.ParseUint(record[3], 10, 64)
		if err != nil {
			id = 99999
		}
		result := db.Where(&models.CBSA{CBSA: uint32(id)}).Find(&cbsas)
		if result.Error != nil {
			fmt.Println(result.Error)
		}
		usage, err := strconv.ParseUint(record[2], 10, 64)
		if err != nil {
			usage = 0
		}
		for _, cbsa := range cbsas {
			if usage != 0 && cbsa.BikeRidership == 0 {
				cbsa.BikeRidership = usage
			}
			db.Save(&cbsa)
		}
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
	db, err := gorm.Open(mysql.Open(url), &gorm.Config{})
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
		"%s:%s@unix(%s)/%s?parseTime=true",
		dbUser,
		dbPass,
		"/cloudsql/"+connectionName,
		dbName,
	)
	db, err := init_db(dbUrl)
	if err != nil {
		log.Fatalln(err)
	}
	// var wg sync.WaitGroup
	// wg.Add(1)
	// db_file, err := models.ReadData(DB_FILE)
	// if err != nil {
	// 	log.Fatalf("Error, file %s could not be read", db_file)
	// }
	// go repopulateGroupTracts(db, db_file, &wg)
	// transit_file, err := models.ReadData(CBSA_TRANSIT_FILE)
	// if err != nil {
	// 	log.Fatalf("Error, file %s could not be read", transit_file)
	// }
	// go addTransitUsage(db, transit_file, &wg)
	// bike_file, err := models.ReadData(CBSA_BIKE_FILE)
	// if err != nil {
	// 	log.Fatalf("Error, file %s could not be read", bike_file)
	// }
	// go addBikeRidership(db, bike_file, &wg)
	// zip_file, err := models.ReadData(ZIPCODE_FILE)
	// if err != nil {
	// 	log.Fatalf("Error, file %s could not be read", zip_file)
	// }
	// go createZipToCBSA(db, zip_file, &wg)

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
