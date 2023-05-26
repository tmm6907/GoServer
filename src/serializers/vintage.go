package serializers

type Vintage struct {
	IsDefault          bool   `json:"isDefault"`
	ID                 string `json:"id"`
	VintageName        string `json:"vintageName"`
	VintageDescription string `json:"vintageDescription"`
}
