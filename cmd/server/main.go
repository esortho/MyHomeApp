package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"myhomeapp/internal/alerting"
	"myhomeapp/internal/api"
	"myhomeapp/internal/config"
	"myhomeapp/internal/db"
	"myhomeapp/internal/handlers"
	"myhomeapp/internal/monitoring"
)

func main() {
	// Parse command line flags
	port := flag.Int("port", 8080, "Port to listen on")
	configPath := flag.String("config", "config.yaml", "Path to the configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	if err := db.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize Aseko client
	asekoClient := api.NewAsekoClient(cfg)
	if err := asekoClient.Initialize(); err != nil {
		log.Fatalf("Failed to initialize Aseko client: %v", err)
	}

	// Initialize alerting service
	alertService, err := initAlertService(cfg)
	if err != nil {
		log.Printf("Warning: Failed to initialize alert service: %v", err)
	} else if alertService != nil && alertService.IsEnabled() {
		log.Println("Alert service initialized successfully")
	}

	// Initialize pool monitor
	poolMonitor, err := monitoring.NewPoolMonitor(asekoClient, alertService, cfg)
	if err != nil {
		log.Printf("Warning: Failed to initialize pool monitor: %v", err)
	} else {
		log.Println("Pool monitor initialized, starting periodic checks...")
		poolMonitor.Start()
	}

	// Set up routes
	http.HandleFunc("/", handlers.IndexHandler)
	http.HandleFunc("/pool", handlers.PoolHandler)
	http.HandleFunc("/measurements", handlers.MeasurementsHandler)
	http.HandleFunc("/lights", handlers.LightsHandler)
	http.HandleFunc("/dashboard", handlers.DashboardHandler)
	http.HandleFunc("/dashboard2", handlers.Dashboard2Handler)
	http.HandleFunc("/history", handlers.HistoryHandler)

	// API endpoints
	http.HandleFunc("/api/measurements", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		measurements, err := asekoClient.GetMeasurements()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get measurements: %v", err), http.StatusInternalServerError)
			return
		}

		// Debug log the measurements
		log.Printf("Raw measurements from Aseko: %+v", measurements)

		// Convert measurements to lowercase property names for frontend
		lowercaseMeasurements := make(map[string]struct {
			Value     float64   `json:"value"`
			Unit      string    `json:"unit"`
			UpdatedAt time.Time `json:"updatedAt"`
		})
		for name, m := range measurements {
			lowercaseMeasurements[name] = struct {
				Value     float64   `json:"value"`
				Unit      string    `json:"unit"`
				UpdatedAt time.Time `json:"updatedAt"`
			}{
				Value:     m.Value,
				Unit:      m.Unit,
				UpdatedAt: m.UpdatedAt,
			}
		}

		// Store measurements in database with consistent names
		for name, m := range measurements {
			// Map Aseko names to our consistent names
			var dbName string
			switch strings.ToUpper(name) {
			case "TEMP":
				dbName = "water_temperature"
			case "PH":
				dbName = "ph"
			case "REDOX":
				dbName = "redox"
			case "WATERFLOW":
				dbName = "water_flow"
			default:
				dbName = strings.ToLower(name)
			}

			if err := db.StoreMeasurement(dbName, m.Value, m.Unit); err != nil {
				log.Printf("Error storing measurement %s: %v", dbName, err)
			} else {
				log.Printf("Stored measurement: name=%s, value=%.2f, unit=%s", dbName, m.Value, m.Unit)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(lowercaseMeasurements)
	})

	http.HandleFunc("/api/measurements/history", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		measurementType := strings.ToLower(r.URL.Query().Get("type"))
		if measurementType == "" {
			http.Error(w, "Missing measurement type", http.StatusBadRequest)
			return
		}

		duration := r.URL.Query().Get("duration")
		if duration == "" {
			duration = "24h" // Default to last 24 hours
		}

		dur, err := time.ParseDuration(duration)
		if err != nil {
			http.Error(w, "Invalid duration", http.StatusBadRequest)
			return
		}

		to := time.Now()
		from := to.Add(-dur)

		log.Printf("Fetching historical measurements: type=%s, from=%v, to=%v", measurementType, from, to)
		measurements, err := db.GetHistoricalMeasurements(measurementType, from, to)
		if err != nil {
			log.Printf("Error getting historical measurements: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		log.Printf("Found %d historical measurements", len(measurements))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(measurements)
	})

	http.HandleFunc("/api/lights", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// TODO: Implement lights endpoint
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]interface{}{})
	})

	http.HandleFunc("/api/lights/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// TODO: Implement light toggle endpoint
		w.WriteHeader(http.StatusOK)
	})

	// API endpoint for pool status
	http.HandleFunc("/api/pool/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if poolMonitor == nil {
			http.Error(w, "Pool monitor not initialized", http.StatusServiceUnavailable)
			return
		}

		status, err := poolMonitor.GetStatus()
		if err != nil {
			log.Printf("Failed to get pool status: %v", err)
			http.Error(w, fmt.Sprintf("Failed to get pool status: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})

	// API endpoint for manual pool check (triggered on page reload)
	http.HandleFunc("/api/pool/check", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if poolMonitor == nil {
			http.Error(w, "Pool monitor not initialized", http.StatusServiceUnavailable)
			return
		}

		// Trigger immediate check
		go poolMonitor.Check()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":     "success",
			"message":    "Pool check triggered",
			"last_check": poolMonitor.GetLastCheckTime(),
		})
	})

	// API endpoint for sending test alerts
	http.HandleFunc("/api/alert/test", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if alertService == nil || !alertService.IsEnabled() {
			http.Error(w, "Alert service is not enabled", http.StatusServiceUnavailable)
			return
		}

		// Get default receivers from config
		var receivers []alerting.Receiver
		for _, r := range cfg.Alerting.DefaultReceivers {
			receivers = append(receivers, alerting.Receiver{
				Email: r.Email,
				Name:  r.Name,
			})
		}

		if len(receivers) == 0 {
			http.Error(w, "No receivers configured", http.StatusBadRequest)
			return
		}

		// Send test alert
		message := alerting.Message{
			Subject:   "MyHomeApp Test Alert",
			Body:      "This is a test alert from MyHomeApp. If you receive this, your alerting is configured correctly!",
			Priority:  alerting.PriorityNormal,
			Timestamp: time.Now(),
		}

		if err := alertService.SendToMultiple(message, receivers); err != nil {
			log.Printf("Failed to send test alert: %v", err)
			http.Error(w, fmt.Sprintf("Failed to send alert: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "Test alert sent successfully",
		})
	})

	// Start server
	log.Printf("Server starting on port %d", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// initAlertService initializes the alerting service from configuration
func initAlertService(cfg *config.Config) (*alerting.AlertService, error) {
	// Check if alerting is configured
	if !cfg.Alerting.Email.Enabled {
		log.Println("Email alerting is disabled in configuration")
		return nil, nil
	}

	// Create email config
	emailConfig := alerting.EmailConfig{
		SMTPHost:     cfg.Alerting.Email.SMTPHost,
		SMTPPort:     cfg.Alerting.Email.SMTPPort,
		SMTPUser:     cfg.Alerting.Email.SMTPUser,
		SMTPPassword: cfg.Alerting.Email.SMTPPassword,
		FromAddress:  cfg.Alerting.Email.FromAddress,
		FromName:     cfg.Alerting.Email.FromName,
		UseTLS:       cfg.Alerting.Email.UseTLS,
		Enabled:      cfg.Alerting.Email.Enabled,
	}

	// Create service config
	serviceConfig := alerting.ServiceConfig{
		EmailConfig: emailConfig,
	}

	return alerting.NewAlertServiceFromConfig(serviceConfig)
}
