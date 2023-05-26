package models

import "gorm.io/gorm"

type Shape struct {
	gorm.Model
	Geoid        uint64 `gorm:"primary_key;ForeignKey:Geoid;"`
	Shape_length float64
	Shape_area   float64
	Geometry     string
}
