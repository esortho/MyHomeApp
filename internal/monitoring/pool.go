package monitoring

import (
	"fmt"
	"log"
	"sync"
	"time"

	"myhomeapp/internal/alerting"
	"myhomeapp/internal/api"
	"myhomeapp/internal/config"
)

// PoolMonitor manages periodic pool monitoring and alerts
type PoolMonitor struct {
	asekoClient      *api.AsekoClient
	alertService     *alerting.AlertService
	config           *config.Config
	checkInterval    time.Duration
	lastCheck        time.Time
	lastCheckMutex   sync.RWMutex
	stopChan         chan struct{}
	defaultReceivers []alerting.Receiver
}

// PoolStatus represents the current pool status
type PoolStatus struct {
	Temperature          float64
	ExpectedTemperature  float64
	TemperatureDelta     float64
	PH                   float64
	Redox                float64
	WaterFlow            string
	WaterFlowAlert       bool
	LastUpdated          time.Time
	TemperatureAlert     bool
	TemperatureAlertType string // "low" or "high"
}

// NewPoolMonitor creates a new pool monitor
func NewPoolMonitor(asekoClient *api.AsekoClient, alertService *alerting.AlertService, cfg *config.Config) (*PoolMonitor, error) {
	// Parse check interval
	interval, err := time.ParseDuration(cfg.Pool.CheckInterval)
	if err != nil {
		return nil, fmt.Errorf("invalid check interval: %w", err)
	}

	// Build default receivers list
	var receivers []alerting.Receiver
	for _, r := range cfg.Alerting.DefaultReceivers {
		receivers = append(receivers, alerting.Receiver{
			Email: r.Email,
			Name:  r.Name,
		})
	}

	return &PoolMonitor{
		asekoClient:      asekoClient,
		alertService:     alertService,
		config:           cfg,
		checkInterval:    interval,
		stopChan:         make(chan struct{}),
		defaultReceivers: receivers,
	}, nil
}

// Start begins the periodic monitoring
func (pm *PoolMonitor) Start() {
	log.Printf("Starting pool monitor with interval: %v", pm.checkInterval)
	
	// Perform initial check
	go pm.Check()

	// Start periodic checks
	ticker := time.NewTicker(pm.checkInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				pm.Check()
			case <-pm.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop stops the periodic monitoring
func (pm *PoolMonitor) Stop() {
	log.Println("Stopping pool monitor")
	close(pm.stopChan)
}

// Check performs an immediate pool check
func (pm *PoolMonitor) Check() {
	pm.lastCheckMutex.Lock()
	pm.lastCheck = time.Now()
	pm.lastCheckMutex.Unlock()

	log.Println("Performing pool check...")

	// Get measurements from Aseko client
	measurements, err := pm.asekoClient.GetMeasurements()
	if err != nil {
		log.Printf("Failed to get measurements: %v", err)
		return
	}

	// Extract temperature
	tempData, hasTempData := measurements["Temp"]
	if !hasTempData {
		log.Println("No temperature data available")
		return
	}

	currentTemp := tempData.Value
	expectedTemp := pm.config.Pool.ExpectedTemperature
	threshold := pm.config.Pool.TemperatureThreshold
	tempDelta := currentTemp - expectedTemp

	log.Printf("Pool check: Current=%.1f°C, Expected=%.1f°C, Delta=%.1f°C, Threshold=±%.1f°C",
		currentTemp, expectedTemp, tempDelta, threshold)

	// Check if alert is needed
	if tempDelta < -threshold {
		// Temperature too low
		pm.sendTemperatureAlert("low", currentTemp, expectedTemp, tempDelta)
	} else if tempDelta > threshold {
		// Temperature too high
		pm.sendTemperatureAlert("high", currentTemp, expectedTemp, tempDelta)
	} else {
		log.Println("Pool temperature is within expected range")
	}

	// Check water flow status
	flowData, hasFlowData := measurements["WaterFlow"]
	if hasFlowData {
		if flowData.Value == 0.0 {
			// Water flow is NO - critical issue
			log.Println("WARNING: Water flow to probes is NO")
			pm.sendWaterFlowAlert()
		} else {
			log.Println("Water flow to probes is OK")
		}
	}
}

// GetStatus returns the current pool status
func (pm *PoolMonitor) GetStatus() (*PoolStatus, error) {
	measurements, err := pm.asekoClient.GetMeasurements()
	if err != nil {
		return nil, fmt.Errorf("failed to get measurements: %w", err)
	}

	status := &PoolStatus{
		ExpectedTemperature: pm.config.Pool.ExpectedTemperature,
		LastUpdated:         time.Now(),
	}

	// Extract temperature
	if tempData, ok := measurements["Temp"]; ok {
		status.Temperature = tempData.Value
		status.TemperatureDelta = status.Temperature - status.ExpectedTemperature
		
		threshold := pm.config.Pool.TemperatureThreshold
		if status.TemperatureDelta < -threshold {
			status.TemperatureAlert = true
			status.TemperatureAlertType = "low"
		} else if status.TemperatureDelta > threshold {
			status.TemperatureAlert = true
			status.TemperatureAlertType = "high"
		}
	}

	// Extract PH
	if phData, ok := measurements["PH"]; ok {
		status.PH = phData.Value
	}

	// Extract Redox
	if redoxData, ok := measurements["Redox"]; ok {
		status.Redox = redoxData.Value
	}

	// Extract Water Flow
	if flowData, ok := measurements["WaterFlow"]; ok {
		if flowData.Value == 1.0 {
			status.WaterFlow = "YES"
			status.WaterFlowAlert = false
		} else {
			status.WaterFlow = "NO"
			status.WaterFlowAlert = true
		}
	}

	return status, nil
}

// GetLastCheckTime returns the time of the last check
func (pm *PoolMonitor) GetLastCheckTime() time.Time {
	pm.lastCheckMutex.RLock()
	defer pm.lastCheckMutex.RUnlock()
	return pm.lastCheck
}

// sendTemperatureAlert sends an alert for temperature deviation
func (pm *PoolMonitor) sendTemperatureAlert(alertType string, current, expected, delta float64) {
	if pm.alertService == nil || !pm.alertService.IsEnabled() {
		log.Println("Alert service not available, skipping temperature alert")
		return
	}

	if len(pm.defaultReceivers) == 0 {
		log.Println("No receivers configured, skipping temperature alert")
		return
	}

	var subject, body string
	var priority alerting.Priority

	if alertType == "low" {
		subject = "Pool Temperature Alert: Below Expected"
		body = fmt.Sprintf(
			"Pool temperature is below expected levels.\n\n"+
				"Current Temperature: %.1f°C\n"+
				"Expected Temperature: %.1f°C\n"+
				"Difference: %.1f°C\n\n"+
				"Time: %s\n\n"+
				"Please check your pool heating system.",
			current, expected, delta, time.Now().Format("2006-01-02 15:04:05"))
		priority = alerting.PriorityHigh
	} else {
		subject = "Pool Temperature Alert: Above Expected"
		body = fmt.Sprintf(
			"Pool temperature is above expected levels.\n\n"+
				"Current Temperature: %.1f°C\n"+
				"Expected Temperature: %.1f°C\n"+
				"Difference: +%.1f°C\n\n"+
				"Time: %s\n\n"+
				"Please check your pool heating system.",
			current, expected, delta, time.Now().Format("2006-01-02 15:04:05"))
		priority = alerting.PriorityNormal
	}

	message := alerting.Message{
		Subject:   subject,
		Body:      body,
		Priority:  priority,
		Timestamp: time.Now(),
	}

	log.Printf("Sending temperature alert: %s", alertType)
	if err := pm.alertService.SendToMultiple(message, pm.defaultReceivers); err != nil {
		log.Printf("Failed to send temperature alert: %v", err)
	} else {
		log.Println("Temperature alert sent successfully")
	}
}

// sendWaterFlowAlert sends a critical alert when water flow is NO
func (pm *PoolMonitor) sendWaterFlowAlert() {
	if pm.alertService == nil || !pm.alertService.IsEnabled() {
		log.Println("Alert service not available, skipping water flow alert")
		return
	}

	if len(pm.defaultReceivers) == 0 {
		log.Println("No receivers configured, skipping water flow alert")
		return
	}

	subject := "CRITICAL: Pool Water Flow Issue"
	body := fmt.Sprintf(
		"CRITICAL ALERT: No water flow detected to pool probes!\n\n"+
			"Status: Water flow is NO\n"+
			"Time: %s\n\n"+
			"IMMEDIATE ACTION REQUIRED:\n"+
			"• Check pool pump operation\n"+
			"• Verify valves are open\n"+
			"• Check for blockages in pipes\n"+
			"• Inspect filter condition\n\n"+
			"No water flow can cause:\n"+
			"• Equipment damage\n"+
			"• Inaccurate readings\n"+
			"• Pool water quality issues\n\n"+
			"Please investigate immediately!",
		time.Now().Format("2006-01-02 15:04:05"))

	message := alerting.Message{
		Subject:   subject,
		Body:      body,
		Priority:  alerting.PriorityCritical,
		Timestamp: time.Now(),
	}

	log.Println("Sending critical water flow alert")
	if err := pm.alertService.SendToMultiple(message, pm.defaultReceivers); err != nil {
		log.Printf("Failed to send water flow alert: %v", err)
	} else {
		log.Println("Water flow alert sent successfully")
	}
}

