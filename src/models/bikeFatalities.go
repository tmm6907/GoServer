package models

import "gorm.io/gorm"

type BikeFatalities struct {
	gorm.Model
	Statefp            uint8 `gorm:"ForeignKey:Statefp;primaryKey;"`
	FatalityPercentage float32
}
