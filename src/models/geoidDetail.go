package models

import "gorm.io/gorm"

type GeoidDetail struct {
	gorm.Model
	Geoid    uint64 `gorm:"ForeignKey:Geoid10;autoIncrement:false"`
	Statefp  uint8
	Countyfp uint16
	Tractce  uint32
	Blkgrpce uint8
}
