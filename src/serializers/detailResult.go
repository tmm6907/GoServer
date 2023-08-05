package serializers

import "encoding/xml"

type DetailResult struct {
	RankID                  uint    `json:"rankID"`
	D2b_e8mixa              float64 `json:"d2b"`
	D2a_ephhm               float64 `json:"d2a"`
	D3b                     float64 `json:"d3b"`
	D4a                     float64 `json:"d4a"`
	PublicTransitEstimate   uint64  `json:"publicTransitEstimate"`
	PublicTransitPercentage float64 `json:"publicTransitPercentage"`
	BikeRidership           uint64  `json:"bikeRidership"`
	BikeRidershipPercentage float64 `json:"bikeRidershipPercentage"`
	BikeRidershipRank       uint8   `json:"bikeRidershipRank"`
	BikePercentageRank      uint8   `json:"bikePercentageRank"`
	BikeFatalityRank        uint8   `json:"bikeFatalityRank"`
	BikeShareRank           uint8   `json:"bikeShareRank"`
}

type DetailResultXML struct {
	XMLName                 xml.Name `xml:"result"`
	RankID                  uint     `xml:"rankID"`
	D2b_e8mixa              float64  `xml:"d2b_e8mixa"`
	D2a_ephhm               float64  `xml:"d2a_ephhm"`
	D3b                     float64  `xml:"d3b"`
	D4a                     float64  `xml:"d4a"`
	PublicTransitEstimate   uint64   `xml:"publicTransitEstimate"`
	PublicTransitPercentage float64  `xml:"publicTransitPercentage"`
	BikeRidership           uint64   `xml:"bikeRidership"`
	BikeRidershipPercentage float64  `xml:"bikeRidershipPercentage"`
	BikeRidershipRank       uint8    `xml:"bikeRidershipRank"`
	BikePercentageRank      uint8    `xml:"bikePercentageRank"`
	BikeFatalityRank        uint8    `xml:"bikeFatalityRank"`
	BikeShareRank           uint8    `xml:"bikeShareRank"`
}
