package models

import "gorm.io/gorm"

type ZipCode struct {
	gorm.Model
	Zipcode string `gorm:"primary_key; index:unique"`
	CBSA    uint32 `gorm:"ForeignKey:CBSA;autoIncrement:false"`
}
