# Pool Monitoring Documentation

## Overview

The pool monitoring system automatically checks your pool temperature and other parameters at regular intervals and sends alerts when values fall outside expected ranges. It combines periodic automated checks with on-demand monitoring triggered by page reloads.

## Features

- ✅ **Periodic Checks**: Automatically checks pool status every hour (configurable)
- ✅ **On-Demand Checks**: Triggered on page reload via API
- ✅ **Temperature Monitoring**: Compares current temperature against expected value
- ✅ **Email Alerts**: Sends notifications when temperature deviates significantly
- ✅ **Configurable Thresholds**: Set your expected temperature and alert threshold

## Configuration

Add the following to your `config.yaml`:

```yaml
pool:
  expected_temperature: 28.0    # Expected pool temperature in °C
  check_interval: "1h"           # How often to check (e.g., "1h", "30m", "2h")
  temperature_threshold: 2.0     # Alert if temperature differs by more than this value
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `expected_temperature` | float | 28.0 | The target pool temperature in degrees Celsius |
| `check_interval` | string | "1h" | How often to perform automatic checks (Go duration format) |
| `temperature_threshold` | float | 2.0 | Temperature deviation threshold for alerts (in °C) |

### Check Interval Examples

- `"30m"` - Check every 30 minutes
- `"1h"` - Check every hour
- `"2h"` - Check every 2 hours
- `"45m"` - Check every 45 minutes

## How It Works

### Automatic Monitoring

1. **Startup**: Pool monitor starts when the application launches
2. **Initial Check**: Performs an immediate check on startup
3. **Periodic Checks**: Runs automatically at the configured interval
4. **Temperature Comparison**: Compares current temperature against expected value
5. **Alert Decision**: Sends alert if temperature is outside threshold range

### Alert Logic

The system sends alerts based on temperature deviation:

```
If (Current Temperature - Expected Temperature) < -Threshold:
    → Send "Temperature Too Low" alert (High Priority)

If (Current Temperature - Expected Temperature) > +Threshold:
    → Send "Temperature Too High" alert (Normal Priority)

Otherwise:
    → No alert (temperature is within acceptable range)
```

#### Example with Default Settings

- **Expected**: 28°C
- **Threshold**: ±2°C
- **Acceptable Range**: 26°C - 30°C

| Current Temp | Result | Alert |
|--------------|--------|-------|
| 25.5°C | Below threshold | ⚠️ Alert: Too Low |
| 26.0°C | At lower bound | ✅ OK |
| 28.0°C | At expected | ✅ OK |
| 30.0°C | At upper bound | ✅ OK |
| 30.5°C | Above threshold | ⚠️ Alert: Too High |

## API Endpoints

### Get Pool Status

Retrieves the current pool status including all measurements and alert status.

```http
GET /api/pool/status
```

**Response:**
```json
{
  "Temperature": 27.5,
  "ExpectedTemperature": 28.0,
  "TemperatureDelta": -0.5,
  "PH": 7.2,
  "Redox": 650.0,
  "WaterFlow": "YES",
  "LastUpdated": "2025-11-03T10:30:00Z",
  "TemperatureAlert": false,
  "TemperatureAlertType": ""
}
```

**Fields:**
- `Temperature`: Current pool temperature in °C
- `ExpectedTemperature`: Configured expected temperature
- `TemperatureDelta`: Difference between current and expected (positive = higher, negative = lower)
- `PH`: Current pH level
- `Redox`: Current Redox level in mV
- `WaterFlow`: Water flow status ("YES" or "NO")
- `LastUpdated`: Timestamp of last measurement
- `TemperatureAlert`: Whether a temperature alert is active
- `TemperatureAlertType`: Type of alert ("low", "high", or empty)

### Trigger Manual Check

Triggers an immediate pool check (useful for page reloads).

```http
POST /api/pool/check
```

**Response:**
```json
{
  "status": "success",
  "message": "Pool check triggered",
  "last_check": "2025-11-03T10:30:00Z"
}
```

## Integration with Frontend

### On Page Load

To trigger a check when the page loads, add this JavaScript:

```javascript
// Trigger pool check on page load
fetch('/api/pool/check', {
  method: 'POST'
})
.then(response => response.json())
.then(data => {
  console.log('Pool check triggered:', data);
})
.catch(error => {
  console.error('Failed to trigger pool check:', error);
});

// Fetch pool status
fetch('/api/pool/status')
.then(response => response.json())
.then(status => {
  console.log('Pool status:', status);
  updatePoolDisplay(status);
})
.catch(error => {
  console.error('Failed to fetch pool status:', error);
});
```

### Display Temperature with Alert Status

```javascript
function updatePoolDisplay(status) {
  const tempElement = document.getElementById('pool-temperature');
  const expectedElement = document.getElementById('expected-temperature');
  const alertElement = document.getElementById('temperature-alert');
  
  tempElement.textContent = status.Temperature.toFixed(1) + '°C';
  expectedElement.textContent = 'Expected: ' + status.ExpectedTemperature.toFixed(1) + '°C';
  
  if (status.TemperatureAlert) {
    alertElement.style.display = 'block';
    alertElement.className = 'alert alert-' + status.TemperatureAlertType;
    
    if (status.TemperatureAlertType === 'low') {
      alertElement.textContent = '⚠️ Temperature is too low';
    } else {
      alertElement.textContent = '⚠️ Temperature is too high';
    }
  } else {
    alertElement.style.display = 'none';
  }
}
```

## Email Alert Format

### Low Temperature Alert

```
Subject: [High] Pool Temperature Alert: Below Expected

Body:
Pool temperature is below expected levels.

Current Temperature: 24.5°C
Expected Temperature: 28.0°C
Difference: -3.5°C

Time: 2025-11-03 10:30:00

Please check your pool heating system.
```

### High Temperature Alert

```
Subject: [Normal] Pool Temperature Alert: Above Expected

Body:
Pool temperature is above expected levels.

Current Temperature: 31.0°C
Expected Temperature: 28.0°C
Difference: +3.0°C

Time: 2025-11-03 10:30:00

Please check your pool heating system.
```

## Monitoring Scenarios

### Scenario 1: Normal Operation

**Configuration:**
- Expected: 28°C
- Threshold: ±2°C
- Interval: 1 hour

**Timeline:**
- 08:00 - Check: 28.5°C → ✅ No alert
- 09:00 - Check: 27.8°C → ✅ No alert
- 10:00 - Check: 28.2°C → ✅ No alert

### Scenario 2: Heating System Failure

**Timeline:**
- 08:00 - Check: 28.0°C → ✅ No alert
- 09:00 - Check: 26.5°C → ✅ No alert (within threshold)
- 10:00 - Check: 25.2°C → ⚠️ Alert sent!
- 11:00 - Check: 24.0°C → (Alert already sent, no duplicate)

### Scenario 3: Page Reload During Issue

**User Action:** Reloads pool page at 10:15
**System Response:**
1. Manual check triggered via API
2. Current temp: 25.0°C
3. Immediate alert sent (if not already sent recently)
4. Page displays warning status

## Best Practices

### Temperature Settings

- **Summer Pools**: 26-28°C typical
- **Heated Pools**: 28-30°C typical
- **Therapy Pools**: 32-34°C typical

### Threshold Settings

- **Relaxed Monitoring**: ±3°C threshold
- **Standard Monitoring**: ±2°C threshold (recommended)
- **Strict Monitoring**: ±1°C threshold

### Check Intervals

- **Active Use**: 30 minutes
- **Standard**: 1 hour (recommended)
- **Low Maintenance**: 2-4 hours

## Troubleshooting

### No Alerts Received

1. **Check alerting configuration**
   ```yaml
   alerting:
     email:
       enabled: true  # Must be true
   ```

2. **Verify default receivers are configured**
   ```yaml
   alerting:
     default_receivers:
       - email: "your-email@example.com"
         name: "Your Name"
   ```

3. **Check logs** for alert sending attempts
   ```bash
   # Look for these log messages:
   # "Sending temperature alert: low" or "high"
   # "Temperature alert sent successfully"
   # "Failed to send temperature alert: ..."
   ```

### Monitor Not Running

Check application logs on startup:
```
Pool monitor initialized, starting periodic checks...
Starting pool monitor with interval: 1h0m0s
Performing pool check...
Pool check: Current=28.0°C, Expected=28.0°C, Delta=0.0°C, Threshold=±2.0°C
Pool temperature is within expected range
```

### Temperature Always Shows Alert

- Verify `expected_temperature` is set correctly for your pool
- Adjust `temperature_threshold` if needed
- Check if pool temperature sensor is working correctly

## Logs and Debugging

The pool monitor provides detailed logging:

```
2025-11-03 10:00:00 Pool monitor initialized, starting periodic checks...
2025-11-03 10:00:00 Starting pool monitor with interval: 1h0m0s
2025-11-03 10:00:01 Performing pool check...
2025-11-03 10:00:02 Pool check: Current=27.5°C, Expected=28.0°C, Delta=-0.5°C, Threshold=±2.0°C
2025-11-03 10:00:02 Pool temperature is within expected range
```

When alerts are triggered:
```
2025-11-03 11:00:01 Performing pool check...
2025-11-03 11:00:02 Pool check: Current=24.5°C, Expected=28.0°C, Delta=-3.5°C, Threshold=±2.0°C
2025-11-03 11:00:02 Sending temperature alert: low
2025-11-03 11:00:03 Temperature alert sent successfully
```

## API Testing

### Test Manual Check

```bash
curl -X POST http://localhost:8080/api/pool/check
```

### Test Status Retrieval

```bash
curl http://localhost:8080/api/pool/status | jq
```

## Advanced Configuration

### Different Expected Temperatures by Time of Day

Currently not supported, but you could implement this by:
1. Adding a time-based configuration
2. Modifying the monitor to use different expected temperatures based on time
3. This would be useful for energy-saving schedules

### Multiple Alert Thresholds

For different priority levels, you could extend the configuration:
```yaml
pool:
  expected_temperature: 28.0
  threshold_warning: 2.0   # Send normal alert
  threshold_critical: 4.0  # Send critical alert
```

## Security Considerations

- Monitor logs may contain temperature data
- Alert emails contain pool status information
- API endpoints should be behind authentication in production
- Consider rate limiting the manual check endpoint

## Performance Notes

- Checks run in background goroutines (non-blocking)
- Manual checks are asynchronous
- Alert sending is fire-and-forget (errors logged but don't block)
- Minimal resource usage (one check per configured interval)

