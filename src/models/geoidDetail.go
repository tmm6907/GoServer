package models

import "gorm.io/gorm"

type GeoidDetail struct {
	gorm.Model
	Geoid     uint64 `gorm:"primary_key;ForeignKey:Geoid;"`
	Statefp   uint8
	Countryfp uint16
	Tractce   uint32
	Blkgrpce  uint8
}
