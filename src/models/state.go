package models

import "gorm.io/gorm"

type State struct {
	gorm.Model
	Name           string
	Statefp        uint8
	BikeFatalities BikeFatalities `gorm:"ForeignKey:Statefp;OnUpdate:CASCADE,OnDelete:SET NULL"`
}
