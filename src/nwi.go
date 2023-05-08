package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	group_tracts "nwi.io/nwi/group_tracts"
)

const CSV_FILE = "Natl_WI.csv"
const ENV_FILE = "./envs/.env"
const RANGE = 500

func crete_entry(db *gorm.DB, data []group_tracts.GroupTract, i int, create_range int) *gorm.DB {
	result := db.Create(data[i:create_range])
	if result.Error != nil {
		log.Fatalln(result.Error)
	}
	return result
}

func repopulateGroupTracts(db *gorm.DB) {
	database, err := group_tracts.ReadData(CSV_FILE)
	if err != nil {
		log.Fatalln(err)
	}
	data := make(chan []group_tracts.GroupTract)
	go func() {
		res := group_tracts.CreateTractGroups(database)
		data <- res
	}()
	res := <-data
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
	)
	return db, nil
}

func main() {
	viper.SetConfigFile(ENV_FILE)
	viper.ReadInConfig()
	port := viper.Get("PORT").(string)
	dbUrl := viper.Get("DB_URL").(string)
	db, err := init_db(dbUrl)
	if err != nil {
		log.Fatalln(err)
	}
	go repopulateGroupTracts(db)
	router := gin.Default()
	group_tracts.RegisterRoutes(router, db)
	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"body": "Hello World!",
		})
	})
	router.Run(port)
}
