package serializers

type AddressScoreResult struct {
	Geoid                          uint64  `json:"geoid"`
	NWI                            float64 `json:"nwi"`
	RegionalTransitUsagePercentage float64 `json:"regionalTransitUsagePercentage"`
	RegionalBikeRidership          uint64  `json:"regionalBikeRidership"`
	SearchedAddress                string  `json:"searchedAddress"`
}
