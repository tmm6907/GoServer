package models

import "gorm.io/gorm"

type CSA struct {
	gorm.Model
	Geoid    uint64 `gorm:"primary_key;ForeignKey:Geoid;"`
	CSA      uint16
	CSA_name string
}
