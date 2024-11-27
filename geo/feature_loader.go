package geo

import (
	_ "embed"
	"encoding/json"
	"log"
)

//go:embed features.json
var feature_file string

func LoadFeatures() []*Feature {
	var features []*Feature

	err := json.Unmarshal([]byte(feature_file), &features)
	if err != nil {
		log.Fatalf("failed to load features: %s\n", err)
	}

	return features
}
