package serializers

type ScoreResult struct {
	ID                             int     `json:"id"`
	Geoid                          uint64  `json:"geoid"`
	CSA_name                       string  `json:"csa_name"`
	CBSA_name                      string  `json:"cbsa_name"`
	NWI                            float64 `json:"nwi"`
	RegionalTransitUsagePercentage float64 `json:"regionalTransitUsagePercentage"`
	RegionalBikeRidership          uint64  `json:"regionalBikeRidership"`
}
