package serializers

type AddressMatch struct {
	TigerLine         TigerLine         `json:"tigerLine"`
	Geographies       Geographies       `json:"geographies"`
	Coordinates       Coordinates       `json:"coordinates"`
	AddressComponents AddressComponents `json:"addressComponents"`
	MatchedAddress    string            `json:"matchedAddress"`
}
