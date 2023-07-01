package serializers

import "encoding/xml"

type AddressScoreResult struct {
	Geoid                          uint64  `json:"geoid"`
	NWI                            float64 `json:"nwi"`
	RegionalTransitUsagePercentage float64 `json:"regionalTransitUsagePercentage"`
	RegionalTransitUsage           uint64  `json:"regionalTransitUsage"`
	RegionalBikeRidership          uint64  `json:"regionalBikeRidership"`
	// TransitScore                   uint8   `json:"transitScore"`
	// BikeScore                      uint8   `json:"bikeScore"`
	SearchedAddress string `json:"searchedAddress"`
}

type AddressScoreResultXML struct {
	XMLName                        xml.Name `xml:"result"`
	Geoid                          uint64   `xml:"geoid"`
	NWI                            float64  `xml:"nwi"`
	RegionalTransitUsagePercentage float64  `xml:"regionalTransitUsagePercentage"`
	RegionalTransitUsage           uint64   `xml:"regionalTransitUsage"`
	RegionalBikeRidership          uint64   `xml:"regionalBikeRidership"`
	// TransitScore                   uint8   `xml:"transitScore"`
	// BikeScore                      uint8   `xml:"bikeScore"`
	SearchedAddress string `xml:"searchedAddress"`
}
