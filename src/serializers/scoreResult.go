package serializers

import "encoding/xml"

type ScoreResults struct {
	ID           int     `json:"id"`
	RankID       uint    `json:"rankID"`
	Geoid        uint64  `json:"geoid"`
	CBSAName     string  `json:"cbsaName"`
	NWI          float64 `json:"nwi"`
	TransitScore uint8   `json:"transitScore"`
	BikeScore    float64 `json:"bikeScore"`
	Format       string  `json:"format"`
}

type ScoreResultsXML struct {
	XMLName      xml.Name `xml:"result"`
	ID           int      `xml:"id,attr"`
	RankID       uint     `xml:"rankID"`
	Geoid        uint64   `xml:"geoid"`
	CBSAName     string   `xml:"cbsaName"`
	NWI          float64  `xml:"nwi"`
	TransitScore uint8    `xml:"transitScore"`
	BikeScore    float64  `xml:"bikeScore"`
	Format       string   `xml:"format"`
}

type XMLResults struct {
	XMLName xml.Name          `xml:"results"`
	Scores  []ScoreResultsXML `xml:"scores"`
}
