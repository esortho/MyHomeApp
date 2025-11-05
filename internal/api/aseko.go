package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"myhomeapp/internal/config"
)

// AsekoClient handles communication with the Aseko API
type AsekoClient struct {
	config       *config.Config
	httpClient   *http.Client
	token        string
	unitList     *UnitListResponse
	selectedUnit *UnitResponse
}

// UnitListResponse represents the response from the unit list query
type UnitListResponse struct {
	Data struct {
		Units struct {
			Cursor string `json:"cursor"`
			Units  []struct {
				Typename     string `json:"__typename"`
				SerialNumber string `json:"serialNumber"`
				Name         string `json:"name"`
				Note         string `json:"note"`
				Online       bool   `json:"online"`
				OfflineFor   int    `json:"offlineFor"`
				HasWarning   bool   `json:"hasWarning"`
				Position     int    `json:"position"`
				BrandName    struct {
					Typename  string `json:"__typename"`
					ID        string `json:"id"`
					Primary   string `json:"primary"`
					Secondary string `json:"secondary"`
				} `json:"brandName"`
				StatusMessages []struct {
					Typename string `json:"__typename"`
					Type     string `json:"type"`
					Severity string `json:"severity"`
					Message  string `json:"message"`
					Detail   string `json:"detail,omitempty"`
				} `json:"statusMessages"`
				Consumables []struct {
					Typename string `json:"__typename"`
					// LiquidConsumable fields
					Canister struct {
						Typename   string `json:"__typename"`
						ID         string `json:"id"`
						HasWarning bool   `json:"hasWarning"`
					} `json:"canister,omitempty"`
					Tube struct {
						Typename   string `json:"__typename"`
						ID         string `json:"id"`
						HasWarning bool   `json:"hasWarning"`
					} `json:"tube,omitempty"`
					// ElectrolyzerConsumable fields
					Electrode struct {
						Typename   string `json:"__typename"`
						HasWarning bool   `json:"hasWarning"`
					} `json:"electrode,omitempty"`
				} `json:"consumables"`
				NotificationConfiguration []struct {
					Typename   string `json:"__typename"`
					ID         string `json:"id"`
					HasWarning bool   `json:"hasWarning"`
				} `json:"notificationConfiguration"`
				UnitModel struct {
					Typename string `json:"__typename"`
					ID       string `json:"id"`
					Tabs     struct {
						HideNotifications bool `json:"hideNotifications"`
						HideConsumables   bool `json:"hideConsumables"`
					} `json:"tabs"`
				} `json:"unitModel"`
			} `json:"units"`
			Typename string `json:"__typename"`
		} `json:"units"`
	} `json:"data"`
}

// UnitResponse represents the response from the unit query
type UnitResponse struct {
	Data struct {
		UnitBySerialNumber struct {
			Typename string `json:"__typename"`
			// Common fields for all types
			SerialNumber string `json:"serialNumber"`
			Name         string `json:"name,omitempty"`
			Note         string `json:"note,omitempty"`
			// Fields specific to Unit and UnitNeverConnected
			StatusMessages []struct {
				Typename string `json:"__typename"`
				Type     string `json:"type"`
				Message  string `json:"message"`
				Severity string `json:"severity"`
				Detail   string `json:"detail"`
			} `json:"statusMessages,omitempty"`
			// Fields specific to Unit
			OfflineFor   string `json:"offlineFor,omitempty"`
			Measurements []struct {
				Typename  string    `json:"__typename"`
				Type      string    `json:"type"`
				Value     float64   `json:"value"`
				Unit      string    `json:"unit"`
				Name      string    `json:"name"`
				UpdatedAt time.Time `json:"updatedAt"`
			} `json:"measurements,omitempty"`
			StatusValues struct {
				Typename string `json:"__typename"`
				Primary  []struct {
					Typename        string `json:"__typename"`
					ID              string `json:"id"`
					Type            string `json:"type"`
					BackgroundColor string `json:"backgroundColor,omitempty"`
					TextColor       string `json:"textColor,omitempty"`
					TopLeft         string `json:"topLeft,omitempty"`
					TopRight        string `json:"topRight,omitempty"`
					Center          struct {
						Typename string `json:"__typename"`
						// StringValue fields
						Value    string `json:"value,omitempty"`
						IconName string `json:"iconName,omitempty"`
						// UpcomingFiltrationPeriodValue fields
						IsNext        bool `json:"isNext,omitempty"`
						Configuration struct {
							Typename             string `json:"__typename"`
							Name                 string `json:"name,omitempty"`
							Speed                int    `json:"speed,omitempty"`
							Start                string `json:"start,omitempty"`
							End                  string `json:"end,omitempty"`
							OverrideIntervalText string `json:"overrideIntervalText,omitempty"`
							PoolFlow             string `json:"poolFlow,omitempty"`
						} `json:"configuration,omitempty"`
					} `json:"center"`
					BottomRight string `json:"bottomRight,omitempty"`
					BottomLeft  struct {
						Typename string `json:"__typename"`
						Prefix   string `json:"prefix,omitempty"`
						Suffix   string `json:"suffix,omitempty"`
						Style    string `json:"style,omitempty"`
					} `json:"bottomLeft,omitempty"`
				} `json:"primary,omitempty"`
				Secondary []struct {
					Typename        string `json:"__typename"`
					ID              string `json:"id"`
					Type            string `json:"type"`
					BackgroundColor string `json:"backgroundColor,omitempty"`
					TextColor       string `json:"textColor,omitempty"`
					TopLeft         string `json:"topLeft,omitempty"`
					TopRight        string `json:"topRight,omitempty"`
					Center          struct {
						Typename string `json:"__typename"`
						// StringValue fields
						Value    string `json:"value,omitempty"`
						IconName string `json:"iconName,omitempty"`
						// UpcomingFiltrationPeriodValue fields
						IsNext        bool `json:"isNext,omitempty"`
						Configuration struct {
							Typename             string `json:"__typename"`
							Name                 string `json:"name,omitempty"`
							Speed                int    `json:"speed,omitempty"`
							Start                string `json:"start,omitempty"`
							End                  string `json:"end,omitempty"`
							OverrideIntervalText string `json:"overrideIntervalText,omitempty"`
							PoolFlow             string `json:"poolFlow,omitempty"`
						} `json:"configuration,omitempty"`
					} `json:"center"`
					BottomRight string `json:"bottomRight,omitempty"`
					BottomLeft  struct {
						Typename string `json:"__typename"`
						Prefix   string `json:"prefix,omitempty"`
						Suffix   string `json:"suffix,omitempty"`
						Style    string `json:"style,omitempty"`
					} `json:"bottomLeft,omitempty"`
				} `json:"secondary,omitempty"`
			} `json:"statusValues,omitempty"`
			Backwash struct {
				Typename      string `json:"__typename"`
				ID            string `json:"id"`
				Running       bool   `json:"running"`
				Duration      int    `json:"duration"`
				Elapsed       int    `json:"elapsed"`
				Configuration struct {
					Typename     string `json:"__typename"`
					OncePerXDays int    `json:"oncePerXDays"`
					Start        string `json:"start"`
					Takes        int    `json:"takes"`
				} `json:"configuration"`
			} `json:"backwash,omitempty"`
			WaterFilling struct {
				Typename                 string `json:"__typename"`
				ID                       string `json:"id"`
				WaterLevel               int    `json:"waterLevel"`
				TotalTime                int    `json:"totalTime"`
				TotalLiters              int    `json:"totalLiters"`
				TotalTimeFromLastReset   int    `json:"totalTimeFromLastReset"`
				TotalLitersFromLastReset int    `json:"totalLitersFromLastReset"`
				LastReset                string `json:"lastReset"`
				LitersPerMinute          int    `json:"litersPerMinute"`
				Configuration            struct {
					Typename       string `json:"__typename"`
					LevelHigh      int    `json:"levelHigh"`
					LevelLow       int    `json:"levelLow"`
					LevelMax       int    `json:"levelMax"`
					LevelMin       int    `json:"levelMin"`
					MaxFillingTime int    `json:"maxFillingTime"`
					Enabled        bool   `json:"enabled"`
				} `json:"configuration"`
			} `json:"waterFilling,omitempty"`
		} `json:"unitBySerialNumber"`
	} `json:"data"`
}

// NewAsekoClient creates a new Aseko API client
func NewAsekoClient(cfg *config.Config) *AsekoClient {
	client := &AsekoClient{
		config:     cfg,
		httpClient: &http.Client{},
	}

	// Initialize in a goroutine to match Java's PostConstruct
	go func() {
		time.Sleep(8 * time.Second) // Wait for authentication to complete
		if err := client.fetchUnitList(); err != nil {
			log.Printf("Error initializing unit list: %v", err)
		}
	}()

	return client
}

func (c *AsekoClient) Login() error {
	loginData := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Cloud    string `json:"cloud"`
	}{
		Email:    c.config.Aseko.Email,
		Password: c.config.Aseko.Password,
		Cloud:    "01HXS50KTV7NRSVNHD617J4CKB",
	}

	jsonData, err := json.Marshal(loginData)
	if err != nil {
		return fmt.Errorf("error marshaling login data: %w", err)
	}

	req, err := http.NewRequest("POST", "https://auth.aseko.acs.aseko.cloud/auth/login", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// Set headers as in the Java implementation
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://aseko.cloud")
	req.Header.Set("Referer", "https://aseko.cloud/")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36")
	req.Header.Set("X-App-Name", "pool-live")
	req.Header.Set("X-App-Version", "4.2.0")
	req.Header.Set("X-Mode", "production")
	req.Header.Set("sec-ch-ua", "\"Not(A:Brand\";v=\"99\", \"Google Chrome\";v=\"133\", \"Chromium\";v=\"133\"")
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", "\"macOS\"")

	// Debug: Print request details
	log.Printf("Login request URL: %s", req.URL.String())
	//log.Printf("Login request headers: %v", req.Header)
	log.Printf("Login request body: %s", string(jsonData))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making login request: %w", err)
	}
	defer resp.Body.Close()

	// Debug: Print response status
	log.Printf("Login response status: %s", resp.Status)

	// Read the response body for debugging
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	// Print the response for debugging
	//log.Printf("Login response body: %s", string(body))

	var loginResp struct {
		User struct {
			ID string `json:"id"`
		} `json:"user"`
		Token string `json:"token"`
	}
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return fmt.Errorf("error decoding login response: %w, response body: %s", err, string(body))
	}

	if loginResp.Token == "" {
		return fmt.Errorf("no token received in response: %s", string(body))
	}

	log.Printf("Login successful, received token: %s...", loginResp.Token[:10])
	c.token = loginResp.Token
	return nil
}

// fetchUnitList fetches the list of units and stores it in the client
func (c *AsekoClient) fetchUnitList() error {
	// Query for the unit list
	unitListQuery := `
		fragment UnitFragment on Unit {
			__typename
			serialNumber
			name
			note
			brandName {
				id
				primary
				secondary
				__typename
			}
			position
			statusMessages {
				__typename
				type
				severity
				message
			}
			consumables {
				__typename
				... on LiquidConsumable {
					canister {
						__typename
						id
						hasWarning
					}
					tube {
						__typename
						id
						hasWarning
					}
					__typename
				}
				... on ElectrolyzerConsumable {
					electrode {
						__typename
						hasWarning
					}
					__typename
				}
			}
			online
			offlineFor
			hasWarning
			notificationConfiguration {
				__typename
				id
				hasWarning
			}
			unitModel {
				__typename
				id
				tabs {
					hideNotifications
					hideConsumables
					__typename
				}
			}
		}

		fragment UnitNeverConnectedFragment on UnitNeverConnected {
			__typename
			serialNumber
			name
			note
			position
			statusMessages {
				__typename
				severity
				type
				message
				detail
			}
		}

		query UnitList($after: String, $first: Int, $search: String) {
			units(after: $after, first: $first, searchQuery: $search) {
				cursor
				units {
					...UnitFragment
					...UnitNeverConnectedFragment
					__typename
				}
				__typename
			}
		}
	`

	reqBody := struct {
		OperationName string `json:"operationName"`
		Query         string `json:"query"`
		Variables     struct {
			After  *string `json:"after"`
			First  int     `json:"first"`
			Search string  `json:"search"`
		} `json:"variables"`
	}{
		OperationName: "UnitList",
		Query:         unitListQuery,
		Variables: struct {
			After  *string `json:"after"`
			First  int     `json:"first"`
			Search string  `json:"search"`
		}{
			After:  nil,
			First:  15,
			Search: "",
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("error marshaling unit query: %w", err)
	}

	req, err := http.NewRequest("POST", c.config.Aseko.BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating unit request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://aseko.cloud")
	req.Header.Set("Referer", "https://aseko.cloud/")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36")
	req.Header.Set("X-App-Name", "pool-live")
	req.Header.Set("X-App-Version", "4.2.0")
	req.Header.Set("X-Mode", "production")
	req.Header.Set("Authorization", "Bearer "+c.token)

	// Debug: Print request details
	log.Printf("Request URL: %s\n", req.URL.String())
	//log.Printf("Request body: %s\n", string(jsonData))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making unit request: %w", err)
	}
	defer resp.Body.Close()

	// Debug: Print response details
	log.Printf("Response status: %s\n", resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading unit response body: %w", err)
	}

	// Debug: Print response body
	//	log.Printf("Response body: %s\n", string(body))

	var unitResponse UnitListResponse
	if err := json.Unmarshal(body, &unitResponse); err != nil {
		return fmt.Errorf("error decoding unit list response: %w, response body: %s", err, string(body))
	}

	c.unitList = &unitResponse
	return nil
}

// SelectUnit selects a unit by its serial number and fetches its details
func (c *AsekoClient) SelectUnit(serialNumber string) error {
	// Query for the specific unit
	unitQuery := `
		fragment StatusValueFragment on StatusValue {
			__typename
			id
			type
			backgroundColor
			textColor
			topLeft
			topRight
			center {
				__typename
				... on StringValue {
					value
					iconName
					__typename
				}
				... on UpcomingFiltrationPeriodValue {
					__typename
					configuration {
						__typename
						name
						speed
						start
						end
						overrideIntervalText
						poolFlow
					}
					isNext
				}
			}
			bottomRight
			bottomLeft {
				__typename
				prefix
				suffix
				style
			}
		}

		fragment BackwashStatusFragment on BackwashStatus {
			__typename
			id
			running
			duration
			elapsed
			configuration {
				__typename
				oncePerXDays
				start
				takes
			}
		}

		fragment StatusMessageFragment on StatusMessage {
			__typename
			type
			severity
			message
			detail
		}

		query UnitDetailStatusQuery($sn: String!) {
			unitBySerialNumber(serialNumber: $sn) {
				__typename
				... on UnitNotFoundError {
					serialNumber
					__typename
				}
				... on UnitAccessDeniedError {
					serialNumber
					__typename
				}
				... on UnitNeverConnected {
					serialNumber
					name
					note
					statusMessages {
						__typename
						type
						message
						severity
						detail
					}
					__typename
				}
				... on Unit {
					__typename
					serialNumber
					name
					note
					statusMessages {
						__typename
						type
						message
						severity
						detail
					}
					offlineFor
					statusValues {
						__typename
						primary {
							...StatusValueFragment
							__typename
						}
						secondary {
							...StatusValueFragment
							__typename
						}
					}
					statusMessages {
						...StatusMessageFragment
						__typename
					}
					backwash {
						...BackwashStatusFragment
						__typename
					}
					waterFilling {
						__typename
						id
						waterLevel
						totalTime
						totalLiters
						totalTimeFromLastReset
						totalLitersFromLastReset
						lastReset
						litersPerMinute
						configuration {
							__typename
							levelHigh
							levelLow
							levelMax
							levelMin
							maxFillingTime
							enabled
						}
					}
				}
			}
		}
	`

	reqBody := struct {
		OperationName string `json:"operationName"`
		Query         string `json:"query"`
		Variables     struct {
			SerialNumber string `json:"sn"`
		} `json:"variables"`
	}{
		OperationName: "UnitDetailStatusQuery",
		Query:         unitQuery,
		Variables: struct {
			SerialNumber string `json:"sn"`
		}{
			SerialNumber: serialNumber,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", c.config.Aseko.BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://aseko.cloud")
	req.Header.Set("Referer", "https://aseko.cloud/")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36")
	req.Header.Set("X-App-Name", "pool-live")
	req.Header.Set("X-App-Version", "4.2.0")
	req.Header.Set("X-Mode", "production")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	// Debug: Print request details
	log.Printf("SelectUnit request URL: %s", req.URL.String())
	//log.Printf("SelectUnit request headers: %v", req.Header)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Debug: Print response status
	log.Printf("SelectUnit response status: %s", resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %v", err)
	}

	var unitResp UnitResponse
	if err := json.Unmarshal(body, &unitResp); err != nil {
		return fmt.Errorf("error parsing response: %v, response body: %s", err, string(body))
	}

	// Check if we have a valid response
	if unitResp.Data.UnitBySerialNumber.SerialNumber == "" {
		// Log the full response for debugging
		responseJSON, _ := json.MarshalIndent(unitResp, "", "  ")
		log.Printf("Empty response for unit %s. Full response: %s", serialNumber, string(responseJSON))

		// Attempt to reconnect and retry
		log.Println("Attempting full reconnect due to empty response...")
		if err := c.Initialize(); err != nil {
			return fmt.Errorf("failed to reconnect: %w", err)
		}
		
		log.Println("Reconnected successfully, retrying SelectUnit...")
		// Retry the unit selection after reconnecting
		return c.selectUnitAfterReconnect(serialNumber)
	}

	// Check the response type and handle accordingly
	switch unitResp.Data.UnitBySerialNumber.Typename {
	case "UnitNotFoundError":
		return fmt.Errorf("unit not found: %s", serialNumber)
	case "UnitAccessDeniedError":
		return fmt.Errorf("access denied for unit: %s", serialNumber)
	case "UnitNeverConnected":
		log.Printf("Unit %s has never been connected", serialNumber)
		c.selectedUnit = &unitResp
		return nil
	case "Unit":
		// Log the status values for debugging
		if unitResp.Data.UnitBySerialNumber.StatusValues.Primary != nil {
			log.Printf("Primary status values:")
			for _, sv := range unitResp.Data.UnitBySerialNumber.StatusValues.Primary {
				log.Printf("  Type: %s", sv.Type)
				if sv.Center.Value != "" {
					log.Printf("    Value: %s", sv.Center.Value)
				}
				if sv.BottomLeft.Prefix != "" || sv.BottomLeft.Suffix != "" {
					log.Printf("    Units: %s%s%s", sv.BottomLeft.Prefix, sv.Center.Value, sv.BottomLeft.Suffix)
				}
			}
		}
		if unitResp.Data.UnitBySerialNumber.StatusValues.Secondary != nil {
			log.Printf("Secondary status values:")
			for _, sv := range unitResp.Data.UnitBySerialNumber.StatusValues.Secondary {
				log.Printf("  Type: %s", sv.Type)
				if sv.Center.Value != "" {
					log.Printf("    Value: %s", sv.Center.Value)
				}
				if sv.BottomLeft.Prefix != "" || sv.BottomLeft.Suffix != "" {
					log.Printf("    Units: %s%s%s", sv.BottomLeft.Prefix, sv.Center.Value, sv.BottomLeft.Suffix)
				}
			}
		}
		c.selectedUnit = &unitResp
		return nil
	case "":
		// If we have a serial number but empty typename, assume it's a Unit
		if unitResp.Data.UnitBySerialNumber.SerialNumber != "" {
			log.Printf("Empty typename but valid serial number, assuming Unit type")
			c.selectedUnit = &unitResp
			return nil
		}

		// Log the full response for debugging
		responseJSON, _ := json.MarshalIndent(unitResp, "", "  ")
		log.Printf("Empty typename in response. Full response: %s", string(responseJSON))
		return fmt.Errorf("empty typename in response for unit: %s", serialNumber)
	default:
		// Log the full response for debugging
		responseJSON, _ := json.MarshalIndent(unitResp, "", "  ")
		log.Printf("Unknown response type: %s. Full response: %s", unitResp.Data.UnitBySerialNumber.Typename, string(responseJSON))
		return fmt.Errorf("unknown response type: %s", unitResp.Data.UnitBySerialNumber.Typename)
	}
}

// GetUnitList returns the cached unit list
func (c *AsekoClient) GetUnitList() *UnitListResponse {
	return c.unitList
}

// GetSelectedUnit returns the cached selected unit
func (c *AsekoClient) GetSelectedUnit() *UnitResponse {
	return c.selectedUnit
}

// Config returns the client's configuration
func (c *AsekoClient) Config() *config.Config {
	return c.config
}

// Token returns the client's authentication token
func (c *AsekoClient) Token() string {
	return c.token
}

// HTTPClient returns the client's HTTP client
func (c *AsekoClient) HTTPClient() *http.Client {
	return c.httpClient
}

// Initialize performs the initial setup of the AsekoClient
func (c *AsekoClient) Initialize() error {
	// Login to get the token
	if err := c.Login(); err != nil {
		return fmt.Errorf("error logging in: %w", err)
	}

	// Fetch the unit list
	if err := c.fetchUnitList(); err != nil {
		return fmt.Errorf("error fetching unit list: %w", err)
	}

	// Select the first unit if available
	if c.unitList != nil && len(c.unitList.Data.Units.Units) > 0 {
		serialNumber := c.unitList.Data.Units.Units[0].SerialNumber
		log.Printf("Attempting to select first unit with serial number: %s", serialNumber)

		if err := c.SelectUnit(serialNumber); err != nil {
			log.Printf("Error selecting first unit: %v", err)

			// Try to refresh the token and retry
			log.Printf("Attempting to refresh token and retry...")
			if err := c.Login(); err != nil {
				return fmt.Errorf("error refreshing token: %w", err)
			}

			// Retry selecting the unit
			if err := c.SelectUnit(serialNumber); err != nil {
				return fmt.Errorf("error selecting first unit after token refresh: %w", err)
			}
		}
	} else {
		log.Printf("No units available to select")
	}

	return nil
}

// GetMeasurements retrieves the current pool measurements
func (c *AsekoClient) GetMeasurements() (map[string]struct {
	Value     float64   `json:"value"`
	Unit      string    `json:"unit"`
	UpdatedAt time.Time `json:"updatedAt"`
}, error) {
	if c.selectedUnit == nil {
		return nil, fmt.Errorf("no unit selected")
	}

	// Refresh the unit data first
	serialNumber := c.selectedUnit.Data.UnitBySerialNumber.SerialNumber
	if err := c.SelectUnit(serialNumber); err != nil {
		return nil, fmt.Errorf("error refreshing unit data: %w", err)
	}

	// Create result map
	result := make(map[string]struct {
		Value     float64   `json:"value"`
		Unit      string    `json:"unit"`
		UpdatedAt time.Time `json:"updatedAt"`
	})

	// Process primary status values
	for _, statusValue := range c.selectedUnit.Data.UnitBySerialNumber.StatusValues.Primary {
		// Map status value type to a consistent key
		var key string
		switch statusValue.Type {
		case "REDOX":
			key = "Redox"
			// Extract value and unit
			value, _ := strconv.ParseFloat(statusValue.Center.Value, 64)
			unit := statusValue.BottomRight
			if unit == "" {
				unit = "mV" // Default unit for Redox
			}
			result[key] = struct {
				Value     float64   `json:"value"`
				Unit      string    `json:"unit"`
				UpdatedAt time.Time `json:"updatedAt"`
			}{
				Value:     value,
				Unit:      unit,
				UpdatedAt: time.Now(), // We don't have the exact update time
			}
		case "PH":
			key = "PH"
			// Extract value
			value, _ := strconv.ParseFloat(statusValue.Center.Value, 64)
			result[key] = struct {
				Value     float64   `json:"value"`
				Unit      string    `json:"unit"`
				UpdatedAt time.Time `json:"updatedAt"`
			}{
				Value:     value,
				Unit:      "",
				UpdatedAt: time.Now(), // We don't have the exact update time
			}
		case "WATER_TEMPERATURE":
			key = "Temp"
			// Extract value and unit
			value, _ := strconv.ParseFloat(statusValue.Center.Value, 64)
			unit := statusValue.BottomRight
			if unit == "" {
				unit = "°C" // Default unit for temperature
			}
			result[key] = struct {
				Value     float64   `json:"value"`
				Unit      string    `json:"unit"`
				UpdatedAt time.Time `json:"updatedAt"`
			}{
				Value:     value,
				Unit:      unit,
				UpdatedAt: time.Now(), // We don't have the exact update time
			}
		case "WATER_FLOW_TO_PROBES":
			key = "WaterFlow"
			// Extract value as string
			value := 0.0
			displayValue := "NO"
			if statusValue.Center.Value == "YES" {
				value = 1.0
				displayValue = "YES"
			}
			result[key] = struct {
				Value     float64   `json:"value"`
				Unit      string    `json:"unit"`
				UpdatedAt time.Time `json:"updatedAt"`
			}{
				Value:     value,
				Unit:      displayValue, // Use YES/NO as the unit
				UpdatedAt: time.Now(),   // We don't have the exact update time
			}
		}
	}

	return result, nil
}

// selectUnitAfterReconnect retries unit selection after a reconnect
// This method does NOT retry again to avoid infinite recursion
func (c *AsekoClient) selectUnitAfterReconnect(serialNumber string) error {
	query := `query UnitDetailStatusQuery($sn: String!) {
		unitBySerialNumber(sn: $sn) {
			__typename
			serialNumber
			name
			note
			statusMessages {
				__typename
				type
				message
				severity
				detail
			}
			offlineFor
			measurements {
				__typename
				type
				value
				unit
				name
				updatedAt
			}
			statusValues {
				__typename
				primary {
					__typename
					id
					type
					backgroundColor
					textColor
					topLeft
					topRight
					center {
						__typename
						value
						iconName
						isNext
						configuration {
							__typename
							name
							speed
							start
							end
							overrideIntervalText
							poolFlow
						}
					}
					bottomRight
					bottomLeft {
						__typename
						prefix
						suffix
						style
					}
				}
				secondary {
					__typename
					id
					type
					backgroundColor
					textColor
					topLeft
					topRight
					center {
						__typename
						value
						iconName
						isNext
						configuration {
							__typename
							name
							speed
							start
							end
							overrideIntervalText
							poolFlow
						}
					}
					bottomRight
					bottomLeft {
						__typename
						prefix
						suffix
						style
					}
				}
			}
			backwash {
				__typename
				id
				running
				duration
				elapsed
				configuration {
					__typename
					oncePerXDays
					start
					takes
				}
			}
			waterFilling {
				__typename
				id
				waterLevel
				totalTime
				totalLiters
				totalTimeFromLastReset
				totalLitersFromLastReset
				lastReset
				litersPerMinute
				configuration {
					__typename
					levelHigh
					levelLow
					levelMax
					levelMin
					maxFillingTime
					enabled
				}
			}
		}
	}`

	variables := map[string]interface{}{
		"sn": serialNumber,
	}

	requestBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", c.config.Aseko.BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-App-Name", "pool-live")
	req.Header.Set("X-App-Version", "4.2.0")
	req.Header.Set("X-Mode", "production")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	log.Printf("Retry after reconnect: SelectUnit request URL: %s", req.URL.String())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("Retry after reconnect: SelectUnit response status: %s", resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %v", err)
	}

	var unitResp UnitResponse
	if err := json.Unmarshal(body, &unitResp); err != nil {
		return fmt.Errorf("error parsing response: %v, response body: %s", err, string(body))
	}

	// Check if we have a valid response (no retry on second attempt)
	if unitResp.Data.UnitBySerialNumber.SerialNumber == "" {
		responseJSON, _ := json.MarshalIndent(unitResp, "", "  ")
		log.Printf("Still empty response after reconnect for unit %s. Full response: %s", serialNumber, string(responseJSON))
		return fmt.Errorf("received empty response for unit %s even after reconnect", serialNumber)
	}

	// Check the response type and handle accordingly
	switch unitResp.Data.UnitBySerialNumber.Typename {
	case "UnitNotFoundError":
		return fmt.Errorf("unit not found: %s", serialNumber)
	case "UnitAccessDeniedError":
		return fmt.Errorf("access denied for unit: %s", serialNumber)
	case "UnitNeverConnected":
		log.Printf("Unit %s has never been connected", serialNumber)
		return fmt.Errorf("unit %s has never been connected", serialNumber)
	}

	c.selectedUnit = &unitResp
	log.Printf("Successfully selected unit after reconnect: %s", serialNumber)
	return nil
}
