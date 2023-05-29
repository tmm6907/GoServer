package api

import (
	"fmt"
	"testing"
)

func TestGetGeoid(t *testing.T) {
	var tests = []struct {
		testString string
		want       string
	}{
		{"1600 Pennsylvania Avenue Northwest Washington DC", "110010062021"},
		{"20 W 34th St., New York, NY 10001", "360610076001"},
		{"272-264 Rice Dr, England, AR 72046", "50850208001"},
		{"12949 LA-10, Pitkin, LA 70656", "221159508003"},
	}
	for _, tt := range tests {
		testname := fmt.Sprint(tt.testString)
		t.Run(testname, func(t *testing.T) {
			ans, err := getGeoid(tt.testString)
			if err != nil {
				fmt.Print(err)
			}
			if ans != tt.want {
				t.Errorf("got %s, want %s", ans, tt.want)
			}
		})
	}
}
