package models

import "time"

type CBSA struct {
	Geoid                   uint64 `gorm:"ForeignKey:Geoid10;primaryKey;"`
	CBSA                    uint32
	CBSA_name               string
	Population              uint64
	PublicTransitEstimate   uint64
	PublicTransitPercentage float64
	BikeRidership           uint64
	BikeRidershipPercentage float64
	CreatedAt               time.Time
	UpdatedAt               time.Time
}
