package models

import "gorm.io/gorm"

type AC struct {
	gorm.Model
	Geoid    uint64 `gorm:"primary_key;ForeignKey:Geoid;"`
	AC_total float64
	AC_water float64
	AC_land  float64
	AC_unpr  float64
}
