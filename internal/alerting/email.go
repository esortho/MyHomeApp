package alerting

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

// EmailConfig holds SMTP configuration
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	FromAddress  string
	FromName     string
	UseTLS       bool
	Enabled      bool
}

// EmailAlerter implements the Alerter interface for email notifications
type EmailAlerter struct {
	config EmailConfig
}

// NewEmailAlerter creates a new email alerter
func NewEmailAlerter(config EmailConfig) *EmailAlerter {
	return &EmailAlerter{
		config: config,
	}
}

// IsEnabled returns whether email alerting is enabled
func (e *EmailAlerter) IsEnabled() bool {
	return e.config.Enabled && e.config.SMTPHost != ""
}

// Send sends an email alert to a single receiver
func (e *EmailAlerter) Send(message Message, receiver Receiver) error {
	if !e.IsEnabled() {
		return fmt.Errorf("email alerting is disabled")
	}

	// Build email content
	from := e.config.FromAddress
	if e.config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", e.config.FromName, e.config.FromAddress)
	}

	to := receiver.Email
	if receiver.Name != "" {
		to = fmt.Sprintf("%s <%s>", receiver.Name, receiver.Email)
	}

	// Format email headers and body
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = fmt.Sprintf("[%s] %s", message.Priority.String(), message.Subject)
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/plain; charset=UTF-8"
	headers["Date"] = message.Timestamp.Format(time.RFC1123Z)

	// Add priority header
	switch message.Priority {
	case PriorityCritical, PriorityHigh:
		headers["X-Priority"] = "1"
		headers["Importance"] = "high"
	case PriorityNormal:
		headers["X-Priority"] = "3"
		headers["Importance"] = "normal"
	case PriorityLow:
		headers["X-Priority"] = "5"
		headers["Importance"] = "low"
	}

	// Build message
	var emailMessage strings.Builder
	for key, value := range headers {
		emailMessage.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	emailMessage.WriteString("\r\n")
	emailMessage.WriteString(message.Body)

	// Send email
	addr := fmt.Sprintf("%s:%s", e.config.SMTPHost, e.config.SMTPPort)
	
	// Set up authentication
	auth := smtp.PlainAuth("", e.config.SMTPUser, e.config.SMTPPassword, e.config.SMTPHost)

	// For port 587, use standard SendMail which handles STARTTLS automatically
	// For port 465, use direct TLS connection
	// smtp.SendMail handles STARTTLS on port 587 automatically when available
	if e.config.SMTPPort == "465" && e.config.UseTLS {
		return e.sendWithTLS(addr, auth, e.config.FromAddress, []string{receiver.Email}, []byte(emailMessage.String()))
	}

	return smtp.SendMail(addr, auth, e.config.FromAddress, []string{receiver.Email}, []byte(emailMessage.String()))
}

// SendToMultiple sends an email alert to multiple receivers
func (e *EmailAlerter) SendToMultiple(message Message, receivers []Receiver) error {
	if !e.IsEnabled() {
		return fmt.Errorf("email alerting is disabled")
	}

	var errors []error
	for _, receiver := range receivers {
		if err := e.Send(message, receiver); err != nil {
			errors = append(errors, fmt.Errorf("failed to send to %s: %w", receiver.Email, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to send to some receivers: %v", errors)
	}

	return nil
}

// sendWithTLS sends email using TLS connection
func (e *EmailAlerter) sendWithTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	// Create TLS configuration
	tlsConfig := &tls.Config{
		ServerName: e.config.SMTPHost,
		MinVersion: tls.VersionTLS12,
	}

	// Connect to SMTP server
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to dial SMTP server: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, e.config.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Authenticate
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	// Set sender
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to add recipient %s: %w", recipient, err)
		}
	}

	// Send message
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to create data writer: %w", err)
	}

	_, err = writer.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return client.Quit()
}

