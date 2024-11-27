package geo

import (
	_ "embed"
	"encoding/csv"
	"log"
	"strconv"
	"strings"
)

//go:embed msoa.csv
var msoa_file string

func loadMSOAs() []*MSOA {
	reader := csv.NewReader(strings.NewReader(msoa_file))
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("failed to load msoas: %s\n", err)
	}

	var msoas []*MSOA
	for _, record := range records {
		popDensity, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			log.Fatalf("invalid population density in msoa source: %v", err)
		}

		msoas = append(msoas, &MSOA{
			Name:              record[0],
			GISCode:           record[1],
			PopulationDensity: popDensity,
		})
	}

	return msoas
}
