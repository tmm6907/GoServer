package serializers

type Input struct {
	Address   AddressQuery `json:"address"`
	Vintage   Vintage      `json:"vintage"`
	Benchmark Benchmark    `json:"benchmark"`
}
