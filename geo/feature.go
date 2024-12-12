package geo

import (
	"encoding/json"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
)

// type Feature map[string]interface{}

type Feature struct {
	Properties map[string]interface{}
	Geometry   geom.T
}

func (feature Feature) Code() string {
	// props := feature["properties"].(map[string]interface{})
	if code, ok := feature.Properties["code"].(string); ok {
		return code
	}

	return ""
}

func (feature Feature) ParentCode() string {
	// props := feature["properties"].(map[string]interface{})
	if parent, ok := feature.Properties["parent"].(string); ok {
		return parent
	}

	return ""
}

func (feature Feature) Name() string {
	// props := feature["properties"].(map[string]interface{})
	if name, ok := feature.Properties["name"].(string); ok {
		return name
	}

	return ""
}

func (feature Feature) Level() string {
	// props := feature["properties"].(map[string]interface{})
	if level, ok := feature.Properties["level"].(string); ok {
		return level
	}

	return ""
}

// UnmarshalJSON implements the custom unmarshalling logic for Command.
func (f *Feature) UnmarshalJSON(data []byte) error {
	// Define an intermediate structure to capture the "type" and raw "payload".
	var intermediate struct {
		Properties map[string]interface{} `json:"properties"`
		Geometry   json.RawMessage        `json:"geometry"`
	}

	// Unmarshal into the intermediate structure.
	if err := json.Unmarshal(data, &intermediate); err != nil {
		return err
	}

	f.Properties = intermediate.Properties

	// Unmarshal GeoJSON directly into a geom.T
	var g geom.T
	if err := geojson.Unmarshal([]byte(intermediate.Geometry), &g); err != nil {
		return err
	}

	f.Geometry = g

	return nil
}
