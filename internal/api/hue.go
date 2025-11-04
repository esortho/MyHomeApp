package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"time"

	"myhomeapp/internal/config"
	"myhomeapp/internal/models"
)

// HueClient handles communication with the Hue Bridge
type HueClient struct {
	config     *config.Config
	httpClient *http.Client
}

// NewHueClient creates a new Hue client
func NewHueClient(config *config.Config) *HueClient {
	return &HueClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetLights retrieves all lights from the Hue Bridge
func (c *HueClient) GetLights() ([]models.HueLight, error) {
	url := fmt.Sprintf("http://%s/api/%s/lights", c.config.Hue.BridgeIP, c.config.Hue.APIKey)
	log.Printf("Attempting to fetch lights from Hue Bridge at: %s", url)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		log.Printf("Error making request to Hue Bridge: %v", err)
		return nil, fmt.Errorf("error fetching lights: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("Hue Bridge response status: %s", resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	log.Printf("Hue Bridge response body: %s", string(body))

	var lights map[string]models.HueLight
	if err := json.Unmarshal(body, &lights); err != nil {
		log.Printf("Error parsing response JSON: %v", err)
		return nil, fmt.Errorf("error parsing lights: %w", err)
	}

	// Convert map to slice and add IDs
	result := make([]models.HueLight, 0, len(lights))
	for id, light := range lights {
		light.ID = id
		// Ensure Reachable is set to true if not present in the response
		if !light.State.Reachable {
			log.Printf("Light %s is not reachable", id)
		}
		result = append(result, light)
	}

	// Sort lights by ID to ensure consistent ordering
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	log.Printf("Successfully retrieved %d lights from Hue Bridge", len(result))
	return result, nil
}

// ToggleLight toggles a light on or off
func (c *HueClient) ToggleLight(lightID string, on bool) error {
	url := fmt.Sprintf("http://%s/api/%s/lights/%s/state", c.config.Hue.BridgeIP, c.config.Hue.APIKey, lightID)
	log.Printf("Attempting to toggle light %s to state: %v", lightID, on)
	log.Printf("Toggle request URL: %s", url)

	// Create request body
	requestBody := struct {
		On bool `json:"on"`
	}{
		On: on,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("error marshaling request body: %w", err)
	}
	log.Printf("Toggle request body: %s", string(jsonData))

	// Create PUT request
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("Toggle response status: %s", resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %w", err)
	}

	log.Printf("Toggle response body: %s", string(body))

	// Check for specific error responses
	if resp.StatusCode != http.StatusOK {
		// Try to parse error response
		var errorResponse []struct {
			Error struct {
				Type        int    `json:"type"`
				Address     string `json:"address"`
				Description string `json:"description"`
			} `json:"error"`
		}

		if err := json.Unmarshal(body, &errorResponse); err == nil && len(errorResponse) > 0 {
			return fmt.Errorf("toggle request failed: %s (type: %d)", errorResponse[0].Error.Description, errorResponse[0].Error.Type)
		}

		return fmt.Errorf("toggle request failed with status %s: %s", resp.Status, string(body))
	}

	log.Printf("Successfully toggled light %s to state: %v", lightID, on)
	return nil
}
