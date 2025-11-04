package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	Aseko struct {
		Email        string `yaml:"email"`
		Password     string `yaml:"password"`
		BaseURL      string `yaml:"base_url"`
		WebSocketURL string `yaml:"websocket_url"`
	} `yaml:"aseko"`
	Pool struct {
		ExpectedTemperature float64 `yaml:"expected_temperature"`
		CheckInterval       string  `yaml:"check_interval"`
		TemperatureThreshold float64 `yaml:"temperature_threshold"`
	} `yaml:"pool"`
	Hue struct {
		BridgeIP string `yaml:"bridge_ip"`
		APIKey   string `yaml:"api_key"`
	} `yaml:"hue"`
	Alerting struct {
		Email struct {
			Enabled      bool   `yaml:"enabled"`
			SMTPHost     string `yaml:"smtp_host"`
			SMTPPort     string `yaml:"smtp_port"`
			SMTPUser     string `yaml:"smtp_user"`
			SMTPPassword string `yaml:"smtp_password"`
			FromAddress  string `yaml:"from_address"`
			FromName     string `yaml:"from_name"`
			UseTLS       bool   `yaml:"use_tls"`
		} `yaml:"email"`
		DefaultReceivers []struct {
			Email string `yaml:"email"`
			Name  string `yaml:"name"`
		} `yaml:"default_receivers"`
	} `yaml:"alerting"`
}

type AsekoConfig struct {
	Email        string `yaml:"email"`
	Password     string `yaml:"password"`
	BaseURL      string `yaml:"base_url"`
	WebSocketURL string `yaml:"websocket_url"`
}

func NewConfig() *Config {
	cfg := &Config{}

	// Default server configuration
	cfg.Server.Port = "8080"

	// Default Aseko API URLs
	cfg.Aseko.BaseURL = "https://graphql.acs.prod.aseko.cloud/graphql"

	// Default pool monitoring configuration
	cfg.Pool.ExpectedTemperature = 28.0 // 28°C
	cfg.Pool.CheckInterval = "1h"       // Check every hour
	cfg.Pool.TemperatureThreshold = 2.0 // Alert if temperature differs by more than 2°C

	return cfg
}

func LoadConfig(path string) (*Config, error) {
	cfg := NewConfig()

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", path)
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Validate required fields
	if cfg.Aseko.Email == "" {
		return nil, fmt.Errorf("aseko.email is required")
	}
	if cfg.Aseko.Password == "" {
		return nil, fmt.Errorf("aseko.password is required")
	}
	if cfg.Hue.BridgeIP == "" {
		return nil, fmt.Errorf("hue.bridge_ip is required")
	}
	if cfg.Hue.APIKey == "" {
		return nil, fmt.Errorf("hue.api_key is required")
	}

	return cfg, nil
}

func GetConfigPath() string {
	// Check for config file in current directory
	if _, err := os.Stat("config.yaml"); err == nil {
		return "config.yaml"
	}

	// Check for config file in home directory
	home, err := os.UserHomeDir()
	if err == nil {
		homeConfig := filepath.Join(home, ".myhomeapp", "config.yaml")
		if _, err := os.Stat(homeConfig); err == nil {
			return homeConfig
		}
	}

	// Default to current directory
	return "config.yaml"
}

// GetEmailAlertingConfig returns the email alerting configuration
func (c *Config) GetEmailAlertingConfig() map[string]interface{} {
	return map[string]interface{}{
		"enabled":       c.Alerting.Email.Enabled,
		"smtp_host":     c.Alerting.Email.SMTPHost,
		"smtp_port":     c.Alerting.Email.SMTPPort,
		"smtp_user":     c.Alerting.Email.SMTPUser,
		"smtp_password": c.Alerting.Email.SMTPPassword,
		"from_address":  c.Alerting.Email.FromAddress,
		"from_name":     c.Alerting.Email.FromName,
		"use_tls":       c.Alerting.Email.UseTLS,
	}
}
