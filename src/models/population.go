package models

import "gorm.io/gorm"

type Population struct {
	gorm.Model
	Geoid     uint64 `gorm:"primary_key;ForeignKey:Geoid;"`
	Total_pop uint16
	CountHU   float64
	HH        float64
	Workers   uint16
}
