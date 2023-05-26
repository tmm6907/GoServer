package serializers

type GeoCodingResultDetail struct {
	Input          Input          `json:"input"`
	AddressMatches AddressMatches `json:"addressMatches"`
}
