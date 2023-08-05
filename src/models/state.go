package models

type State struct {
	Name           string
	Statefp        uint8          `gorm:"primary_key; auto_increment:false"`
	BikeFatalities BikeFatalities `gorm:"ForeignKey:Statefp;OnUpdate:CASCADE,OnDelete:SET NULL"`
}
