# Alerting Module Documentation

## Overview

The alerting module provides a simple internal API for sending notifications via email. It's designed to be extensible, allowing for additional notification channels to be added in the future.

## Architecture

The module consists of:

- **Core API**: Defines `Message` and `Receiver` structures
- **Alerter Interface**: Allows multiple notification channels
- **Email Implementation**: SMTP-based email notifications
- **Alert Service**: Manages multiple alerters and routing

## Configuration

Add the following to your `config.yaml`:

```yaml
alerting:
  email:
    enabled: true
    smtp_host: "smtp.gmail.com"
    smtp_port: "587"
    smtp_user: "your-email@gmail.com"
    smtp_password: "your-app-password"
    from_address: "your-email@gmail.com"
    from_name: "MyHomeApp Alerts"
    use_tls: true
  default_receivers:
    - email: "recipient@example.com"
      name: "Recipient Name"
```

### Email Provider Examples

#### Gmail
- **SMTP Host**: `smtp.gmail.com`
- **SMTP Port**: `587` (TLS) or `465` (SSL)
- **Note**: You need to use an [App Password](https://support.google.com/accounts/answer/185833) instead of your regular password

#### Outlook/Office 365
- **SMTP Host**: `smtp.office365.com`
- **SMTP Port**: `587`
- **Use TLS**: `true`

#### Custom SMTP Server
- Configure with your server's SMTP settings
- Supports both TLS and non-TLS connections

## Internal API Usage

### Basic Message Structure

```go
import "myhomeapp/internal/alerting"

message := alerting.Message{
    Subject:   "Pool Temperature Alert",
    Body:      "Pool temperature is below threshold: 20°C",
    Priority:  alerting.PriorityHigh,
    Timestamp: time.Now(),
}
```

### Priority Levels

- `PriorityLow`: Low priority notifications
- `PriorityNormal`: Normal priority (default)
- `PriorityHigh`: High priority alerts
- `PriorityCritical`: Critical alerts requiring immediate attention

### Receiver Structure

```go
receiver := alerting.Receiver{
    Email: "user@example.com",
    Name:  "John Doe",
}
```

### Sending Alerts

#### Send to a Single Receiver

```go
err := alertService.Send(message, receiver)
if err != nil {
    log.Printf("Failed to send alert: %v", err)
}
```

#### Send to Multiple Receivers

```go
receivers := []alerting.Receiver{
    {Email: "admin@example.com", Name: "Admin"},
    {Email: "user@example.com", Name: "User"},
}

err := alertService.SendToMultiple(message, receivers)
if err != nil {
    log.Printf("Failed to send alerts: %v", err)
}
```

#### Send to Default Receivers

```go
// Using default receivers from config
var receivers []alerting.Receiver
for _, r := range cfg.Alerting.DefaultReceivers {
    receivers = append(receivers, alerting.Receiver{
        Email: r.Email,
        Name:  r.Name,
    })
}

err := alertService.SendToMultiple(message, receivers)
```

## Testing Alerts

The application includes a test endpoint to verify your alerting configuration:

```bash
curl -X POST http://localhost:8080/api/alert/test
```

Expected response:
```json
{
  "status": "success",
  "message": "Test alert sent successfully"
}
```

## Example Use Cases

### Pool Temperature Alert

```go
temperature := 18.5
threshold := 20.0

if temperature < threshold {
    message := alerting.Message{
        Subject:   "Pool Temperature Alert",
        Body:      fmt.Sprintf("Pool temperature (%.1f°C) is below threshold (%.1f°C)", temperature, threshold),
        Priority:  alerting.PriorityHigh,
        Timestamp: time.Now(),
    }
    
    err := alertService.SendToMultiple(message, defaultReceivers)
    if err != nil {
        log.Printf("Failed to send temperature alert: %v", err)
    }
}
```

### System Status Alert

```go
message := alerting.Message{
    Subject:   "MyHomeApp Status Report",
    Body:      "System is operating normally. All sensors are online.",
    Priority:  alerting.PriorityNormal,
    Timestamp: time.Now(),
}

err := alertService.SendToMultiple(message, defaultReceivers)
```

### Critical Equipment Failure

```go
message := alerting.Message{
    Subject:   "CRITICAL: Equipment Failure Detected",
    Body:      "Pool pump has stopped responding. Immediate attention required.",
    Priority:  alerting.PriorityCritical,
    Timestamp: time.Now(),
}

err := alertService.SendToMultiple(message, adminReceivers)
```

## Extending the Module

### Adding New Alert Channels

To add a new notification channel (e.g., SMS, Slack, Push notifications):

1. Implement the `Alerter` interface:

```go
type NewAlerter struct {
    config NewConfig
}

func (a *NewAlerter) Send(message Message, receiver Receiver) error {
    // Implementation
}

func (a *NewAlerter) SendToMultiple(message Message, receivers []Receiver) error {
    // Implementation
}

func (a *NewAlerter) IsEnabled() bool {
    // Implementation
}
```

2. Update `AlertService` to include the new alerter
3. Add configuration structure to `config.yaml`
4. Update the initialization in `main.go`

## Security Considerations

- **Store credentials securely**: Never commit `config.yaml` with real credentials to version control
- **Use App Passwords**: For Gmail and similar services, use app-specific passwords
- **TLS/SSL**: Always use TLS when possible (`use_tls: true`)
- **Validate inputs**: The module validates email addresses and message content
- **Rate limiting**: Consider implementing rate limiting to prevent alert spam

## Troubleshooting

### Email Not Sending

1. **Check SMTP credentials**: Verify username and password are correct
2. **App Password**: If using Gmail, ensure you're using an App Password, not your regular password
3. **Firewall**: Ensure port 587 (or 465) is not blocked
4. **TLS Settings**: Try toggling `use_tls` setting
5. **Check logs**: Review application logs for detailed error messages

### Gmail-Specific Issues

- Enable "Less secure app access" or use an App Password
- Verify 2-factor authentication is configured
- Check for blocks in your Google Account security settings

## API Reference

### Types

#### Message
```go
type Message struct {
    Subject   string
    Body      string
    Timestamp time.Time
    Priority  Priority
}
```

#### Receiver
```go
type Receiver struct {
    Email string
    Name  string
}
```

#### Priority
```go
type Priority int

const (
    PriorityLow Priority = iota
    PriorityNormal
    PriorityHigh
    PriorityCritical
)
```

### Interfaces

#### Alerter
```go
type Alerter interface {
    Send(message Message, receiver Receiver) error
    SendToMultiple(message Message, receivers []Receiver) error
    IsEnabled() bool
}
```

### AlertService Methods

- `Send(message Message, receiver Receiver) error`: Send alert to single receiver
- `SendToMultiple(message Message, receivers []Receiver) error`: Send alert to multiple receivers
- `IsEnabled() bool`: Check if service is enabled
- `Enable()`: Enable the alert service
- `Disable()`: Disable the alert service

