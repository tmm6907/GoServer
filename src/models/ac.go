package models

import "gorm.io/gorm"

type AC struct {
	gorm.Model
	Geoid    uint64 `gorm:"ForeignKey:Geoid10;primaryKey;"`
	AC_total float64
	AC_water float64
	AC_land  float64
	AC_unpr  float64
}
