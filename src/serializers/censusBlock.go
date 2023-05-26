package serializers

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
