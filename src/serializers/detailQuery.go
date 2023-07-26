package serializers

type DetailQuery struct {
	Fields string `form:"fields" json:"fields" xml:"fields"`
	Format string `form:"format" json:"format" xml:"format"`
}

func (q *DetailQuery) SetFormat() {
	if q.Format == "" {
		q.Format = "json"
	}
}
