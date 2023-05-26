package models

import (
	"gorm.io/gorm"
)

type BlockGroup struct {
	gorm.Model
	Geoid10     uint64 `gorm:"primary_key;"`
	Geoid20     uint64
	GeoidDetail GeoidDetail `gorm:"ForeignKey:Geoid"`
	CSA         CSA         `gorm:"ForeignKey:Geoid"`
	CBSA        CBSA        `gorm:"ForeignKey:Geoid"`
	AC          AC          `gorm:"ForeignKey:Geoid"`
	Population  Population  `gorm:"ForeignKey:Geoid"`
	Rank        Rank        `gorm:"ForeignKey:Geoid"`
	Shape       Shape       `gorm:"ForeignKey:Geoid"`
}
