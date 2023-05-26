package models

import "gorm.io/gorm"

type CBSA struct {
	gorm.Model
	Geoid             uint64 `gorm:"primary_key;ForeignKey:Geoid;"`
	CBSA              uint32
	CBSA_name         string
	PublicTansitUsage float64 `gorm:"default:0"`
	BikeRidership     uint64  `gorm:"default:0"`
}
