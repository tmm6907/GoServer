package serializers

import "encoding/xml"

type DetailResult struct {
	RankID                  uint    `json:"rankID"`
	D2b_e8mixa              float64 `json:"d2b"`
	D2a_ephhm               float64 `json:"d2a"`
	D3b                     float64 `json:"d3b"`
	D4a                     float64 `json:"d4a"`
	D2a_ranked              float32 `json:"d2aRanked"`
	D2b_ranked              float32 `json:"d2bRanked"`
	D3b_ranked              float32 `json:"d3bRanked"`
	D4a_ranked              float32 `json:"d4aRanked"`
	PublicTransitEstimate   uint64  `json:"publicTransitEstimate"`
	PublicTransitPercentage float64 `json:"publicTransitPercentage"`
	BikeRidership           uint64  `json:"bikeRidership"`
	BikeRidershipPercentage float64 `json:"bikeRidershipPercentage"`
}

type DetailResultXML struct {
	XMLName                 xml.Name `xml:"result"`
	RankID                  uint     `xml:"rankID"`
	D2b_e8mixa              float64  `xml:"d2b_e8mixa"`
	D2a_ephhm               float64  `xml:"d2a_ephhm"`
	D3b                     float64  `xml:"d3b"`
	D4a                     float64  `xml:"d4a"`
	D2a_ranked              float32  `xml:"d2aRanked"`
	D2b_ranked              float32  `xml:"d2bRanked"`
	D3b_ranked              float32  `xml:"d3bRanked"`
	D4a_ranked              float32  `xml:"d4aRanked"`
	PublicTransitEstimate   uint64   `xml:"publicTransitEstimate"`
	PublicTransitPercentage float64  `xml:"publicTransitPercentage"`
	BikeRidership           uint64   `xml:"bikeRidership"`
	BikeRidershipPercentage float64  `xml:"bikeRidershipPercentage"`
}
