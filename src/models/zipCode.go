package models

import "gorm.io/gorm"

type ZipCode struct {
	gorm.Model
	Zipcode string `gorm:"primary_key"`
	CBSA    uint32 `gorm:"ForeignKey:CBSA"`
}
