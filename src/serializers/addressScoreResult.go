package serializers

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
