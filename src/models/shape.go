package models

import "gorm.io/gorm"

type Shape struct {
	gorm.Model
	Geoid        uint64 `gorm:"ForeignKey:Geoid10;primaryKey;"`
	Shape_length float64
	Shape_area   float64
}
