package serializers

import "encoding/xml"

type AddressScoreResult struct {
	Geoid           uint64  `json:"geoid"`
	CBSAName        string  `json:"cbsaName"`
	NWI             float64 `json:"nwi"`
	TransitScore    uint8   `json:"transitScore"`
	BikeScore       uint8   `json:"bikeScore"`
	SearchedAddress string  `json:"searchedAddress"`
	Format          string  `json:"format"`
}

type AddressScoreResultXML struct {
	XMLName         xml.Name `xml:"result"`
	Geoid           uint64   `xml:"geoid"`
	CBSAName        string   `json:"cbsaName"`
	NWI             float64  `xml:"nwi"`
	TransitScore    uint8    `xml:"transitScore"`
	BikeScore       uint8    `xml:"bikeScore"`
	SearchedAddress string   `xml:"searchedAddress"`
	Format          string   `xml:"format"`
}
