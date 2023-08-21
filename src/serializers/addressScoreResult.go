package serializers

import "encoding/xml"

type AddressScoreResult struct {
	RankID          uint    `json:"rankID"`
	Geoid           uint64  `json:"geoid"`
	NWI             float64 `json:"nwi"`
	TransitScore    uint8   `json:"transitScore"`
	BikeScore       float64 `json:"bikeScore"`
	SearchedAddress string  `json:"searchedAddress"`
	Format          string  `json:"format"`
}

type AddressScoreResultXML struct {
	XMLName         xml.Name `xml:"result"`
	RankID          uint     `xml:"rankID"`
	Geoid           uint64   `xml:"geoid"`
	NWI             float64  `xml:"nwi"`
	TransitScore    uint8    `xml:"transitScore"`
	BikeScore       float64  `xml:"bikeScore"`
	SearchedAddress string   `xml:"searchedAddress"`
	Format          string   `xml:"format"`
}
