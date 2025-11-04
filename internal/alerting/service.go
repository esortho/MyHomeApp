package alerting

import (
	"fmt"
)

// ServiceConfig holds the configuration for the alert service
type ServiceConfig struct {
	EmailConfig      EmailConfig
	DefaultReceivers []Receiver
}

// NewAlertServiceFromConfig creates a new alert service from configuration
func NewAlertServiceFromConfig(cfg ServiceConfig) (*AlertService, error) {
	// Create email alerter if configured
	var emailAlerter Alerter
	if cfg.EmailConfig.Enabled {
		if cfg.EmailConfig.SMTPHost == "" {
			return nil, fmt.Errorf("SMTP host is required when email alerting is enabled")
		}
		if cfg.EmailConfig.SMTPPort == "" {
			return nil, fmt.Errorf("SMTP port is required when email alerting is enabled")
		}
		if cfg.EmailConfig.FromAddress == "" {
			return nil, fmt.Errorf("from address is required when email alerting is enabled")
		}

		emailAlerter = NewEmailAlerter(cfg.EmailConfig)
	}

	return NewAlertService(emailAlerter), nil
}

// GetDefaultReceivers returns the configured default receivers
func (s *AlertService) GetDefaultReceivers() []Receiver {
	return []Receiver{}
}

// SendToDefault sends an alert to the default receivers
func (s *AlertService) SendToDefault(message Message, defaultReceivers []Receiver) error {
	if len(defaultReceivers) == 0 {
		return fmt.Errorf("no default receivers configured")
	}

	return s.SendToMultiple(message, defaultReceivers)
}

