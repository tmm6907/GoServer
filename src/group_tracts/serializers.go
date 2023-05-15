package group_tracts

type ScoresQuery struct {
	Limit  int
	Offset int
}

type Address struct {
	Address string `json:"address"`
}

type Vintage struct {
	IsDefault          bool   `json:"isDefault"`
	ID                 string `json:"id"`
	VintageName        string `json:"vintageName"`
	VintageDescription string `json:"vintageDescription"`
}

type Benchmark struct {
	IsDefault            bool   `json:"isDefault"`
	BenchmarkDescription string `json:"benchmarkDescription"`
	ID                   string `json:"id"`
	BenchmarkName        string `json:"benchmarkName"`
}

type Input struct {
	Address   Address   `json:"address"`
	Vintage   Vintage   `json:"vintage"`
	Benchmark Benchmark `json:"benchmark"`
}
type TigerLine struct {
	Side        string `json:"side"`
	TigerLineID string `json:"tigerLineId"`
}

type CensusBlock struct {
	Suffix    string `json:"SUFFIX"`
	Pop100    int    `json:"POP100"`
	Geoid     string `json:"GEOID"`
	Cenlat    string `json:"CENTLAT"`
	Block     string `json:"BLOCK"`
	Areawater int    `json:"AREAWATER"`
	State     string `json:"STATE"`
	Basename  string `json:"BASENAME"`
	OID       int
	LSADC     string
	INTPTLAT  string
	FUNCSTAT  string
	NAME      string
	OBJECTID  int
	TRACT     string
	CENTLON   string
	BLKGRP    string
	AREALAND  int
	HU100     int
	INTPTLON  string
	MTFCC     string
	LWBLKTYP  string
	UR        string
	COUNTY    string
}

type StateLDUpper []CensusBlock
type CensusBlocks []CensusBlock

type Geographies struct {
	CensusBlocks `json:"Census Blocks"`
}

type Coordinates struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

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

type AddressMatch struct {
	TigerLine         TigerLine         `json:"tigerLine"`
	Geographies       Geographies       `json:"geographies"`
	Coordinates       Coordinates       `json:"coordinates"`
	AddressComponents AddressComponents `json:"addressComponents"`
	MatchedAddress    string            `json:"matchedAddress"`
}
type AddressMatches []AddressMatch

type GeoCodingResultDetail struct {
	Input          Input          `json:"input"`
	AddressMatches AddressMatches `json:"addressMatches"`
}

type GeoCodingResult struct {
	Result GeoCodingResultDetail `json:"result"`
}

type AddressScoreResult struct {
	Geoid                          uint64  `json:"geoid"`
	NWI                            float64 `json:"nwi"`
	RegionalTransitUsagePercentage float64 `json:"regionalTransitUsagePercentage"`
	RegionalBikeRidership          uint64  `json:"regionalBikeRidership"`
	SearchedAddress                string  `json:"searchedAddress"`
}

type ScoreResult struct {
	ID                             int     `json:"id"`
	Geoid                          uint64  `json:"geoid"`
	CSA_name                       string  `json:"csa_name"`
	CBSA_name                      string  `json:"cbsa_name"`
	NWI                            float64 `json:"nwi"`
	RegionalTransitUsagePercentage float64 `json:"regionalTransitUsagePercentage"`
	RegionalBikeRidership          uint64  `json:"regionalBikeRidership"`
}

type ScoreResults []ScoreResult

type ZipcodeResult struct {
	Zipcode string `json:"zipcode"`
}
