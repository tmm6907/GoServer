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
	group_tracts "nwi.io/nwi/group_tracts"
)

const DB_FILE = "gs://open-nwi/Natl_WI.csv"
const CBSA_TRANSIT_FILE = "gs://open-nwi/CBSA_Public_Transit_Usage.csv"
const CBSA_BIKE_FILE = "gs://open-nwi/CBSA_Bicylce_Ridership.csv"
const ZIPCODE_FILE string = "gs://open-nwi/zip07_cbsa06.csv"
const RANGE = 500

func crete_entry(db *gorm.DB, data []group_tracts.GroupTract, i int, create_range int) *gorm.DB {
	result := db.Create(data[i:create_range])
	if result.Error != nil {
		log.Fatalln(result.Error)
	}
	return result
}

func crete_zipcode_entry(db *gorm.DB, data []group_tracts.Zipcode, i int, create_range int) *gorm.DB {
	result := db.Create(data[i:create_range])
	if result.Error != nil {
		log.Fatalln(result.Error)
	}
	return result
}

func addTransitUsage(db *gorm.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	var cbsas []group_tracts.CBSA
	transit_data, transit_err := group_tracts.ReadData(CBSA_TRANSIT_FILE)
	if transit_err != nil {
		log.Fatalln(transit_err)
	}
	for _, record := range transit_data {
		result := db.Where("cbsa=?", record[4]).Find(&cbsas)
		if result.Error != nil {
			log.Fatalln(result.Error)
		}
		usage, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			fmt.Println(err)
		}
		for _, cbsa := range cbsas {
			cbsa.PublicTansitUsage = usage
			db.Save(&cbsa)
		}
	}

}

func addBikeRidership(db *gorm.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	var cbsas []group_tracts.CBSA
	bike_data, transit_err := group_tracts.ReadData(CBSA_BIKE_FILE)
	if transit_err != nil {
		log.Fatalln(transit_err)
	}
	for _, record := range bike_data {
		result := db.Where("cbsa=?", record[3]).Find(&cbsas)
		if result.Error != nil {
			log.Fatalln(result.Error)
		}
		usage, err := strconv.ParseUint(record[2], 10, 64)
		if err != nil {
			fmt.Println(err)
		}
		for _, cbsa := range cbsas {
			cbsa.BikeRidership = usage
			db.Save(&cbsa)
		}
	}

}
func createZipToCBSA(db *gorm.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	zipcodes := group_tracts.MatchZipToCBSA(ZIPCODE_FILE)
	data_len := len(zipcodes)
	for i := 0; i < data_len; i += RANGE {
		if i+RANGE < data_len {
			result := crete_zipcode_entry(db, zipcodes, i, i+RANGE)
			if result.Error != nil {
				log.Fatal(result.Error)
			}
		}
	}
}

func repopulateGroupTracts(db *gorm.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	database, db_err := group_tracts.ReadData(DB_FILE)
	if db_err != nil {
		log.Fatalln(db_err)
	}
	db_data := make(chan []group_tracts.GroupTract)
	go func() {
		res := group_tracts.CreateTractGroups(database)
		db_data <- res
	}()
	res := <-db_data
	data_len := len(res)
	for i := 0; i < data_len; i += RANGE {
		if i+RANGE < data_len {
			result := crete_entry(db, res, i, i+RANGE)
			if result.Error != nil {
				log.Fatal(result.Error)
			}
		}
	}
	remainder := data_len % RANGE
	if remainder > 0 {
		result := crete_entry(db, res, data_len-remainder, data_len)
		if result.Error != nil {
			log.Fatalln(result.Error)
		}
	}
}

func init_db(url string) (*gorm.DB, error) {
	// Initialize

	db, err := gorm.Open(mysql.Open(url), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(
		&group_tracts.GroupTract{},
		&group_tracts.GeoidDetail{},
		&group_tracts.CSA{},
		&group_tracts.CBSA{},
		&group_tracts.AC{},
		&group_tracts.Population{},
		&group_tracts.Rank{},
		&group_tracts.Shape{},
		&group_tracts.Zipcode{},
	)
	return db, nil
}

func main() {
	port := os.Getenv("INTERNAL_PORT")
	if port == "" {
		log.Fatalf("Fatal Error in connect_unix.go: %s environment variable not set.", port)
	}
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
	var wg sync.WaitGroup
	wg.Add(3)
	go repopulateGroupTracts(db, &wg)
	// go addCBSA_Usage(db, &wg)
	go addBikeRidership(db, &wg)
	go createZipToCBSA(db, &wg)
	router := gin.Default()
	group_tracts.RegisterRoutes(router, db)
	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"body": "Hello World!",
		})
	})
	router.Run(port)
	wg.Wait()
}
