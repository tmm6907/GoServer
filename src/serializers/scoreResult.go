package serializers

import "encoding/xml"

type ScoreResult struct {
	ID                             int     `json:"id"`
	Geoid                          uint64  `json:"geoid"`
	CSA_name                       string  `json:"csa_name"`
	CBSA_name                      string  `json:"cbsa_name"`
	NWI                            float64 `json:"nwi"`
	RegionalTransitUsagePercentage float64 `json:"regionalTransitUsagePercentage"`
	RegionalTransitUsage           uint64  `json:"regionalTransitUsage"`
	RegionalBikeRidership          uint64  `json:"regionalBikeRidership"`
	// TransitScore                   uint8   `json:"transitScore"`
	// BikeScore                      uint8   `json:"bikeScore"`
}

type ScoreResultXML struct {
	XMLName                        xml.Name `xml:"result"`
	ID                             int      `xml:"id,attr"`
	Geoid                          uint64   `xml:"geoid"`
	CSA_name                       string   `xml:"csa_name"`
	CBSA_name                      string   `xml:"cbsa_name"`
	NWI                            float64  `xml:"nwi"`
	RegionalTransitUsagePercentage float64  `xml:"regionalTransitUsagePercentage"`
	RegionalTransitUsage           uint64   `xml:"regionalTransitUsage"`
	RegionalBikeRidership          uint64   `xml:"regionalBikeRidership"`
	// TransitScore                   uint8   `xml:"transitScore"`
	// BikeScore                      uint8   `xml:"bikeScore"`
}
