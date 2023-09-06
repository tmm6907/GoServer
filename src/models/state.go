package models

import "time"

type State struct {
	Name           string
	Statefp        uint8          `gorm:"primaryKey;autoIncrement:false;"`
	BikeFatalities BikeFatalities `gorm:"ForeignKey:Statefp;OnUpdate:CASCADE,OnDelete:SET NULL"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
