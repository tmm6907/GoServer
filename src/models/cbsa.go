package models

import "gorm.io/gorm"

type CBSA struct {
	gorm.Model
	Geoid                   uint64 `gorm:"ForeignKey:Geoid10;autoIncrement:false"`
	CBSA                    uint32
	CBSA_name               string
	Population              uint64  `gorm:"default:0"`
	PublicTransitEstimate   uint64  `gorm:"default:0"`
	PublicTransitPercentage float64 `gorm:"default:0"`
	BikeRidership           uint64  `gorm:"default:0"`
	BikeRidershipPercentage float64 `gorm:"default:0"`
}
