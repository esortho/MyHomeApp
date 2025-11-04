package alerting

import (
	"fmt"
	"time"
)

// Message represents an alert message to be sent
type Message struct {
	Subject   string
	Body      string
	Timestamp time.Time
	Priority  Priority
}

// Priority defines the importance level of an alert
type Priority int

const (
	PriorityLow Priority = iota
	PriorityNormal
	PriorityHigh
	PriorityCritical
)

func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "Low"
	case PriorityNormal:
		return "Normal"
	case PriorityHigh:
		return "High"
	case PriorityCritical:
		return "Critical"
	default:
		return "Unknown"
	}
}

// Receiver represents a destination for alerts
type Receiver struct {
	Email string
	Name  string
}

// Alerter defines the interface for sending alerts
type Alerter interface {
	Send(message Message, receiver Receiver) error
	SendToMultiple(message Message, receivers []Receiver) error
	IsEnabled() bool
}

// AlertService manages multiple alert channels
type AlertService struct {
	emailAlerter Alerter
	enabled      bool
}

// NewAlertService creates a new alert service
func NewAlertService(emailAlerter Alerter) *AlertService {
	return &AlertService{
		emailAlerter: emailAlerter,
		enabled:      true,
	}
}

// Send sends an alert using the appropriate channel
func (s *AlertService) Send(message Message, receiver Receiver) error {
	if !s.enabled {
		return fmt.Errorf("alert service is disabled")
	}

	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	// Send via email if available and receiver has email
	if s.emailAlerter != nil && s.emailAlerter.IsEnabled() && receiver.Email != "" {
		if err := s.emailAlerter.Send(message, receiver); err != nil {
			return fmt.Errorf("failed to send email alert: %w", err)
		}
	}

	return nil
}

// SendToMultiple sends an alert to multiple receivers
func (s *AlertService) SendToMultiple(message Message, receivers []Receiver) error {
	if !s.enabled {
		return fmt.Errorf("alert service is disabled")
	}

	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	var errors []error
	for _, receiver := range receivers {
		if err := s.Send(message, receiver); err != nil {
			errors = append(errors, fmt.Errorf("failed to send to %s: %w", receiver.Email, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to send alerts: %v", errors)
	}

	return nil
}

// Enable enables the alert service
func (s *AlertService) Enable() {
	s.enabled = true
}

// Disable disables the alert service
func (s *AlertService) Disable() {
	s.enabled = false
}

// IsEnabled returns whether the alert service is enabled
func (s *AlertService) IsEnabled() bool {
	return s.enabled
}

