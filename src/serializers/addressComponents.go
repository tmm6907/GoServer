package serializers

type AddressComponents struct {
	Zip             string `json:"zip"`
	StreetName      string `json:"streetName"`
	PreType         string `json:"preType"`
	City            string `json:"city"`
	PreDirection    string `json:"preDirection"`
	SuffixDirection string `json:"suffixDirection"`
	FromAddress     string `json:"fromAddress"`
	State           string `json:"state"`
	SuffixType      string `json:"suffixType"`
	ToAddress       string `json:"toAddress"`
	SuffixQualifier string `json:"suffixQualifier"`
	PreQualifier    string `json:"preQualifier"`
}
