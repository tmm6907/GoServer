package group_tracts

import (
	"time"

	"gorm.io/gorm"
)

type GroupTract struct {
	Geoid10     uint64 `gorm:"primary_key;autoIncrement:false;uniqueIndex;"`
	Geoid20     uint64
	GeoidDetail GeoidDetail `gorm:"ForeignKey:Geoid"`
	CSA         CSA         `gorm:"ForeignKey:Geoid;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CBSA        CBSA        `gorm:"ForeignKey:Geoid;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	AC          AC          `gorm:"ForeignKey:Geoid;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Population  Population  `gorm:"ForeignKey:Geoid;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Rank        Rank        `gorm:"ForeignKey:Geoid;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Shape       Shape       `gorm:"ForeignKey:Geoid;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type GeoidDetail struct {
	gorm.Model
	Geoid     uint64 `gorm:"primary_key;ForeignKey:Geoid;"`
	Statefp   uint8
	Countryfp uint16
	Tractce   uint32
	Blkgrpce  uint8
}

type CSA struct {
	gorm.Model
	Geoid    uint64 `gorm:"primary_key;ForeignKey:Geoid;"`
	CSA      uint16
	CSA_name string
}

type CBSA struct {
	gorm.Model
	Geoid             uint64 `gorm:"primary_key;ForeignKey:Geoid;"`
	CBSA              uint32
	CBSA_name         string
	PublicTansitUsage float64 `gorm:"default:0"`
	BikeRidership     uint64  `gorm:"default:0"`
}

type AC struct {
	gorm.Model
	Geoid    uint64 `gorm:"primary_key;ForeignKey:Geoid;"`
	AC_total float64
	AC_water float64
	AC_land  float64
	AC_unpr  float64
}

type Population struct {
	gorm.Model
	Geoid     uint64 `gorm:"primary_key;ForeignKey:Geoid;"`
	Total_pop uint16
	CountHU   float64
	HH        float64
	Workers   uint16
}

type Rank struct {
	gorm.Model
	Geoid      uint64 `gorm:"primary_key;ForeignKey:Geoid;"`
	D2b_e8mixa float64
	D2a_ephhm  float64
	D3b        float64
	D4a        float64
	D2a_ranked float32
	D2b_ranked float32
	D3b_ranked float32
	D4a_ranked float32
	NWI        float64
}

type Shape struct {
	gorm.Model
	Geoid        uint64 `gorm:"primary_key;ForeignKey:Geoid;"`
	Shape_length float64
	Shape_area   float64
	Geometry     string
}
type Zipcode struct {
	Zipcode string `gorm:"primary_key"`
	CBSA    uint32 `gorm:"ForeignKey:CBSA"`
}
