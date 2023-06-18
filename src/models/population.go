package models

import "gorm.io/gorm"

type Population struct {
	gorm.Model
	Geoid     uint64 `gorm:"ForeignKey:Geoid10;autoIncrement:false"`
	Total_pop uint16
	CountHU   float64
	HH        float64
	Workers   uint16
}
