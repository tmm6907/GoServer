package models

import "gorm.io/gorm"

type BikeShare struct {
	gorm.Model
	FIPS       uint32
	Year       uint16
	Count      uint16
	Percentage float32
}
