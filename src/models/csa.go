package models

import "time"

type CSA struct {
	Geoid     uint64 `gorm:"ForeignKey:Geoid10;primaryKey;"`
	CSA       uint16
	CSA_name  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
