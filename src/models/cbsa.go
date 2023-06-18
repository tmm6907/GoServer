package models

import "gorm.io/gorm"

type CBSA struct {
	gorm.Model
	Geoid                  uint64 `gorm:"ForeignKey:Geoid10;autoIncrement:false"`
	CBSA                   uint32
	CBSA_name              string
	Population             uint64  `gorm:"default:0"`
	PublicTansitEstimate   uint64  `gorm:"default:0"`
	PublicTansitPercentage float64 `gorm:"default:0"`
	BikeRidership          uint64  `gorm:"default:0"`
}
