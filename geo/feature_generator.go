package geo

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

func fetchFeatures(gis_codes []string) ([]Feature, error) {
	const numWorkers = 20

	tasks := make(chan string, len(gis_codes)) // Channel to distribute tasks

	features := make(chan Feature, len(gis_codes))
	errors := make(chan error, len(gis_codes))

	var wg sync.WaitGroup

	// Worker function
	worker := func() {
		defer wg.Done()
		for gis_code := range tasks {
			data, err := fetchFeature(gis_code)
			if err != nil {
				errors <- fmt.Errorf("error fetching data for %s: %v", gis_code, err)
				continue
			}

			features <- data
		}
	}

	// Start workers
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go worker()
	}

	// Distribute tasks
	for _, gis_code := range gis_codes {
		tasks <- gis_code
	}
	close(tasks)

	// Wait for workers to complete
	wg.Wait()

	close(errors)
	close(features)

	// Collect errors
	if len(errors) > 0 {
		for err := range errors {
			fmt.Println(err) // Log errors (optional)
		}
		return nil, fmt.Errorf("some fetch operations failed, check logs")
	}

	results := make([]Feature, 0, len(gis_codes))
	for feature := range features {
		results = append(results, feature)
	}

	return results, nil
}

func fetchFeature(gisCode string) (Feature, error) {
	// debugging
	log.Printf("fetching geo data for MSOA code: %s\n", gisCode)

	url := fmt.Sprintf("https://findthatpostcode.uk/areas/%s.geojson", gisCode)
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 response: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var data Feature
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	log.Printf("successfully fetched MSOA data MSOA code: %s\n", gisCode)

	feature := data["features"].([]interface{})[0].(map[string]interface{})
	return feature, nil
}

// saveJson serializes the msoas to JSON and saves them to a file.
func saveJson(data interface{}, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Pretty-print JSON with indentation
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to write JSON to file: %v", err)
	}

	return nil
}

func GenerateFeatures(path string) {
	msoas := loadMSOAs()

	gis_codes := make([]string, 0, len(msoas))
	for _, msoa := range msoas {
		gis_codes = append(gis_codes, msoa.GISCode)
	}

	features, err := fetchFeatures(gis_codes)
	if err != nil {
		fmt.Printf("Error fetching data for msoas: %v\n", err)
		return
	}

	parent_map := make(map[string]interface{}, 0)
	for _, feature := range features {
		// tag the feature type
		feature["properties"].(map[string]interface{})["level"] = "msoa"

		parent := feature.ParentCode()
		if parent != "" {
			log.Printf("found parent for %s: %s", feature.Code(), parent)
			parent_map[parent] = nil
		}
	}

	parents := make([]string, 0, len(parent_map))
	for parent := range parent_map {
		parents = append(parents, parent)
	}

	parent_features, err := fetchFeatures(parents)
	if err != nil {
		fmt.Printf("Error fetching data for msoas: %v\n", err)
		return
	}

	for _, feature := range parent_features {
		// tag the feature type
		feature["properties"].(map[string]interface{})["level"] = "lad"

		// append parent to features
		features = append(features, feature)
	}

	saveJson(features, path)
}
