package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"sync"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/file"
	applog "google.golang.org/appengine/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	group_tracts "nwi.io/nwi/group_tracts"
)

var db_file string
var cbsa_transit_file string
var cbsa_bike_file string
var zipcode_file string

const RANGE = 500

type Bucket struct {
	Client     *storage.Client
	BucketName string
	Bucket     *storage.BucketHandle

	W       io.Writer
	Ctx     context.Context
	CleanUp []string
	Failed  bool
}

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

func addTransitUsage(db *gorm.DB, database [][]string, wg *sync.WaitGroup) {
	defer wg.Done()
	var cbsas []group_tracts.CBSA

	for _, record := range database {
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

func addBikeRidership(db *gorm.DB, database [][]string, wg *sync.WaitGroup) {
	defer wg.Done()
	var cbsas []group_tracts.CBSA
	for _, record := range database {
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
func createZipToCBSA(db *gorm.DB, database [][]string, wg *sync.WaitGroup) {
	defer wg.Done()
	zipcodes := group_tracts.MatchZipToCBSA(database)
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

func repopulateGroupTracts(db *gorm.DB, database [][]string, wg *sync.WaitGroup) {
	defer wg.Done()
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

// func addTransitUsage(db *gorm.DB, database [][]string, wg *sync.WaitGroup){
// 	defer wg.Done()
// 	var results []group_tracts.CBSA
// 	for _, record := range database {
// 		db.Where("cbsa=?", record[4]).Find(&results)
// 		percentage, err := strconv.ParseFloat(record[2], 64)
// 		if err != nil{
// 			log.Fatalln(err)
// 		}
// 		for _, item := range results {
// 			item.PublicTansitUsage = percentage
// 		}
// 		db.Save(&results)
// 	}
// }

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

	router := gin.Default()
	group_tracts.RegisterRoutes(router, db)
	router.GET("/", func(ctx *gin.Context) {
		a_ctx := appengine.NewContext(ctx.Request)
		bucket, err := file.DefaultBucketName(ctx)
		if err != nil {
			applog.Errorf(ctx, "failed to get default GCS bucket name: %v", err)
		}
		client, err := storage.NewClient(ctx)
		if err != nil {
			applog.Errorf(ctx, "failed to create client: %v", err)
			return
		}
		defer client.Close()
		fmt.Printf("Demo GCS Application running from Version: %v\n", appengine.VersionID(a_ctx))
		buf := &bytes.Buffer{}
		b := &Bucket{
			W:          buf,
			Ctx:        ctx,
			Client:     client,
			Bucket:     client.Bucket("open-nwi"),
			BucketName: bucket,
		}
		var wg sync.WaitGroup
		wg.Add(4)
		db_file = "Natl_WI.csv"
		cbsa_transit_file = "CBSA_Public_Transit_Usage.csv"
		cbsa_bike_file = "CBSA_Bicylce_Ridership.csv"
		zipcode_file = "zip07_cbsa06.csv"
		file, err := b.readFile(db_file)
		if err != nil {
			log.Fatalln(err)
		}
		go repopulateGroupTracts(db, file, &wg)
		transit_file, err := b.readFile(cbsa_transit_file)
		if err != nil {
			log.Fatalln(err)
		}
		go addTransitUsage(db, transit_file, &wg)
		bike_file, err := b.readFile(cbsa_bike_file)
		if err != nil {
			log.Fatalln(err)
		}
		go addBikeRidership(db, bike_file, &wg)
		zip_file, err := b.readFile(cbsa_bike_file)
		if err != nil {
			log.Fatalln(err)
		}
		go createZipToCBSA(db, zip_file, &wg)

		ctx.JSON(200, gin.H{
			"body": "Hello World!",
		})
		wg.Wait()
	})
	router.Run(port)
}
func (b *Bucket) errorf(format string, args ...interface{}) {
	b.Failed = true
	fmt.Fprintln(b.W, fmt.Sprintf(format, args...))
	applog.Errorf(b.Ctx, format, args...)
}

func (b *Bucket) readFile(fileName string) ([][]string, error) {
	rc, err := b.Bucket.Object(fileName).NewReader(b.Ctx)
	if err != nil {
		b.errorf("readFile: unable to open file from bucket %q, file %q: %v", b.BucketName, fileName, err)
	}
	r := csv.NewReader(rc)
	// skip first line
	if _, err := r.Read(); err != nil {
		return [][]string{}, err
	}

	records, err := r.ReadAll()

	if err != nil {
		return [][]string{}, err
	}

	return records, nil
}
