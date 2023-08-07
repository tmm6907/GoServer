package models

import "gorm.io/gorm"

type Rank struct {
	gorm.Model
	Geoid              uint64 `gorm:"ForeignKey:Geoid10;autoIncrement:false;index:unique"`
	D2b_e8mixa         float64
	D2a_ephhm          float64
	D3b                float64
	D4a                float64
	D2a_ranked         float32
	D2b_ranked         float32
	D3b_ranked         float32
	D4a_ranked         float32
	BikeCountRank      uint8 `gorm:"default:1"`
	BikePercentageRank uint8 `gorm:"default:1"`
	BikeFatalityRank   uint8
	BikeShareRank      uint8 `gorm:"default:1"`
	NWI                float64
	TransitScore       uint8   `gorm:"default:0"`
	BikeScore          float64 `gorm:"default:0"`
}
