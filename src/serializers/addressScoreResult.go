package serializers

import "encoding/xml"

type AddressScoreResult struct {
	Geoid                          uint64  `json:"geoid"`
	NWI                            float64 `json:"nwi"`
	RegionalTransitUsagePercentage float64 `json:"regionalTransitUsagePercentage"`
	RegionalTransitUsage           uint64  `json:"regionalTransitUsage"`
	RegionalBikeRidership          uint64  `json:"regionalBikeRidership"`
	SearchedAddress                string  `json:"searchedAddress"`
	Format                         string  `json:"format"`
}

type AddressScoreResultXML struct {
	XMLName                        xml.Name `xml:"result"`
	Geoid                          uint64   `xml:"geoid"`
	NWI                            float64  `xml:"nwi"`
	RegionalTransitUsagePercentage float64  `xml:"regionalTransitUsagePercentage"`
	RegionalTransitUsage           uint64   `xml:"regionalTransitUsage"`
	RegionalBikeRidership          uint64   `xml:"regionalBikeRidership"`
	SearchedAddress                string   `xml:"searchedAddress"`
	Format                         string   `xml:"format"`
}
