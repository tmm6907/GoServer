package serializers

type Benchmark struct {
	IsDefault            bool   `json:"isDefault"`
	BenchmarkDescription string `json:"benchmarkDescription"`
	ID                   string `json:"id"`
	BenchmarkName        string `json:"benchmarkName"`
}
