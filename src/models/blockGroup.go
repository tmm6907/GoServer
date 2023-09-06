package models

import "time"

type BlockGroup struct {
	Geoid10     uint64 `gorm:"primaryKey; auto_increment:false"`
	Geoid20     uint64
	GeoidDetail GeoidDetail `gorm:"ForeignKey:Geoid;OnUpdate:CASCADE,OnDelete:SET NULL"`
	CSA         CSA         `gorm:"ForeignKey:Geoid;OnUpdate:CASCADE,OnDelete:SET NULL"`
	CBSA        CBSA        `gorm:"ForeignKey:Geoid;OnUpdate:CASCADE,OnDelete:SET NULL"`
	AC          AC          `gorm:"ForeignKey:Geoid;OnUpdate:CASCADE,OnDelete:SET NULL"`
	Population  Population  `gorm:"ForeignKey:Geoid;OnUpdate:CASCADE,OnDelete:SET NULL"`
	Rank        Rank        `gorm:"ForeignKey:Geoid;OnUpdate:CASCADE,OnDelete:SET NULL"`
	Shape       Shape       `gorm:"ForeignKey:Geoid;OnUpdate:CASCADE,OnDelete:SET NULL"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
