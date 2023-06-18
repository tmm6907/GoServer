package models

import "gorm.io/gorm"

type CSA struct {
	gorm.Model
	Geoid    uint64 `gorm:"ForeignKey:Geoid10;autoIncrement:false"`
	CSA      uint16
	CSA_name string
}
