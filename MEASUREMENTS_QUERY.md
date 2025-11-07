# Aseko Cloud GraphQL Measurements Query - Detailed Documentation

## Overview

This document describes in detail how to query the Aseko Cloud API for pool measurements, the structure of the response, and how to extract and process the values.

## GraphQL Query Structure

### 1. The Query

The main query used to fetch measurements is `UnitDetailStatusQuery`:

```graphql
query UnitDetailStatusQuery($sn: String!) {
  unitBySerialNumber(serialNumber: $sn) {
    __typename
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
```

### 2. Required Fragment: StatusValueFragment

This fragment defines the structure for status values (where measurements are stored):

```graphql
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
```

### 3. Request Variables

```json
{
  "sn": "UNIT_SERIAL_NUMBER"
}
```

### 4. Request Body

The complete HTTP POST request body:

```json
{
  "operationName": "UnitDetailStatusQuery",
  "query": "... (full query with fragments) ...",
  "variables": {
    "sn": "UNIT_SERIAL_NUMBER"
  }
}
```

### 5. Required HTTP Headers

```http
POST https://graphql.acs.prod.aseko.cloud/graphql
Content-Type: application/json
Accept: */*
Accept-Language: en
Connection: keep-alive
Origin: https://aseko.cloud
Referer: https://aseko.cloud/
Sec-Fetch-Dest: empty
Sec-Fetch-Mode: cors
Sec-Fetch-Site: same-site
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36
X-App-Name: pool-live
X-App-Version: 4.2.0
X-Mode: production
Authorization: Bearer YOUR_AUTH_TOKEN
```

## Response Structure

### Top-Level Response

```json
{
  "data": {
    "unitBySerialNumber": {
      "__typename": "Unit",
      "serialNumber": "ABC123",
      "name": "Pool",
      "note": "",
      "statusValues": {
        "__typename": "StatusValues",
        "primary": [ ... ],
        "secondary": [ ... ]
      },
      "statusMessages": [ ... ],
      "backwash": { ... },
      "waterFilling": { ... }
    }
  }
}
```

### StatusValues Structure

**This is where the measurements are stored!**

The `statusValues.primary` array contains the main pool measurements. Each item has this structure:

```json
{
  "__typename": "StatusValue",
  "id": "unique-id",
  "type": "WATER_TEMPERATURE",
  "backgroundColor": "#00A3E0",
  "textColor": "#FFFFFF",
  "topLeft": "Pool",
  "topRight": "",
  "center": {
    "__typename": "StringValue",
    "value": "28.5",
    "iconName": null
  },
  "bottomRight": "°C",
  "bottomLeft": {
    "__typename": "TextStyle",
    "prefix": "",
    "suffix": "",
    "style": ""
  }
}
```

## Measurement Types and Extraction

### 1. Water Temperature (`WATER_TEMPERATURE`)

**Type**: `"WATER_TEMPERATURE"`

**Value Location**: `statusValue.center.value` (as string)

**Unit Location**: `statusValue.bottomRight` (e.g., "°C")

**Example**:
```json
{
  "type": "WATER_TEMPERATURE",
  "center": { "value": "28.5" },
  "bottomRight": "°C"
}
```

**Extraction**:
```go
if statusValue.Type == "WATER_TEMPERATURE" {
    value, _ := strconv.ParseFloat(statusValue.Center.Value, 64)  // 28.5
    unit := statusValue.BottomRight                                 // "°C"
    if unit == "" {
        unit = "°C" // Default
    }
}
```

### 2. pH Level (`PH`)

**Type**: `"PH"`

**Value Location**: `statusValue.center.value` (as string)

**Unit**: No unit (pH is unitless)

**Example**:
```json
{
  "type": "PH",
  "center": { "value": "7.2" },
  "bottomRight": ""
}
```

**Extraction**:
```go
if statusValue.Type == "PH" {
    value, _ := strconv.ParseFloat(statusValue.Center.Value, 64)  // 7.2
    // pH has no unit
}
```

### 3. Redox/ORP (`REDOX`)

**Type**: `"REDOX"`

**Value Location**: `statusValue.center.value` (as string)

**Unit Location**: `statusValue.bottomRight` (e.g., "mV")

**Example**:
```json
{
  "type": "REDOX",
  "center": { "value": "650" },
  "bottomRight": "mV"
}
```

**Extraction**:
```go
if statusValue.Type == "REDOX" {
    value, _ := strconv.ParseFloat(statusValue.Center.Value, 64)  // 650.0
    unit := statusValue.BottomRight                                 // "mV"
    if unit == "" {
        unit = "mV" // Default
    }
}
```

### 4. Water Flow (`WATER_FLOW_TO_PROBES`)

**Type**: `"WATER_FLOW_TO_PROBES"`

**Value Location**: `statusValue.center.value` (as string: "YES" or "NO")

**Unit**: N/A (boolean status)

**Example**:
```json
{
  "type": "WATER_FLOW_TO_PROBES",
  "center": { "value": "YES" },
  "bottomRight": ""
}
```

**Extraction**:
```go
if statusValue.Type == "WATER_FLOW_TO_PROBES" {
    value := 0.0
    displayValue := "NO"
    if statusValue.Center.Value == "YES" {
        value = 1.0
        displayValue = "YES"
    }
    // Store value as 1.0 for YES, 0.0 for NO
    // Store displayValue as "YES" or "NO"
}
```

## Complete Value Extraction Process

### Step-by-Step Algorithm

1. **Make GraphQL Request**
   - POST to `https://graphql.acs.prod.aseko.cloud/graphql`
   - Include authentication token
   - Send query with unit serial number

2. **Parse JSON Response**
   ```go
   var response UnitResponse
   json.Unmarshal(body, &response)
   ```

3. **Navigate to Primary Status Values**
   ```go
   statusValues := response.Data.UnitBySerialNumber.StatusValues.Primary
   ```

4. **Iterate Through Status Values**
   ```go
   for _, statusValue := range statusValues {
       // Process each measurement
   }
   ```

5. **Extract Based on Type**
   ```go
   switch statusValue.Type {
   case "WATER_TEMPERATURE":
       // Extract temperature
   case "PH":
       // Extract pH
   case "REDOX":
       // Extract redox
   case "WATER_FLOW_TO_PROBES":
       // Extract water flow status
   }
   ```

6. **Parse String Values to Numbers**
   ```go
   value, err := strconv.ParseFloat(statusValue.Center.Value, 64)
   ```

7. **Extract Units**
   ```go
   unit := statusValue.BottomRight
   // Or from BottomLeft.Prefix/Suffix if present
   ```

## Example Response Data

### Complete Example for Temperature

```json
{
  "__typename": "StatusValue",
  "id": "temp-123",
  "type": "WATER_TEMPERATURE",
  "backgroundColor": "#00A3E0",
  "textColor": "#FFFFFF",
  "topLeft": "Pool",
  "topRight": "",
  "center": {
    "__typename": "StringValue",
    "value": "28.5",
    "iconName": null
  },
  "bottomRight": "°C",
  "bottomLeft": {
    "__typename": "TextStyle",
    "prefix": "",
    "suffix": "",
    "style": ""
  }
}
```

**Extracted Result**:
- Type: `WATER_TEMPERATURE`
- Value: `28.5` (float)
- Unit: `°C`
- Display: "28.5°C"

### Complete Example for Water Flow

```json
{
  "__typename": "StatusValue",
  "id": "flow-456",
  "type": "WATER_FLOW_TO_PROBES",
  "backgroundColor": "#00A3E0",
  "textColor": "#FFFFFF",
  "topLeft": "Water Flow",
  "topRight": "",
  "center": {
    "__typename": "StringValue",
    "value": "NO",
    "iconName": null
  },
  "bottomRight": "",
  "bottomLeft": {
    "__typename": "TextStyle",
    "prefix": "",
    "suffix": "",
    "style": ""
  }
}
```

**Extracted Result**:
- Type: `WATER_FLOW_TO_PROBES`
- Value: `0.0` (numeric representation)
- Unit: `NO` (display value)
- Status: NO FLOW (critical condition)

## Data Structure in Go

### Response Structure

```go
type UnitResponse struct {
    Data struct {
        UnitBySerialNumber struct {
            Typename     string `json:"__typename"`
            SerialNumber string `json:"serialNumber"`
            Name         string `json:"name,omitempty"`
            StatusValues struct {
                Typename  string `json:"__typename"`
                Primary   []StatusValue `json:"primary,omitempty"`
                Secondary []StatusValue `json:"secondary,omitempty"`
            } `json:"statusValues,omitempty"`
        } `json:"unitBySerialNumber"`
    } `json:"data"`
}

type StatusValue struct {
    Typename        string `json:"__typename"`
    ID              string `json:"id"`
    Type            string `json:"type"`
    BackgroundColor string `json:"backgroundColor,omitempty"`
    TextColor       string `json:"textColor,omitempty"`
    TopLeft         string `json:"topLeft,omitempty"`
    TopRight        string `json:"topRight,omitempty"`
    Center          struct {
        Typename string `json:"__typename"`
        Value    string `json:"value,omitempty"`
        IconName string `json:"iconName,omitempty"`
    } `json:"center"`
    BottomRight string `json:"bottomRight,omitempty"`
    BottomLeft  struct {
        Typename string `json:"__typename"`
        Prefix   string `json:"prefix,omitempty"`
        Suffix   string `json:"suffix,omitempty"`
        Style    string `json:"style,omitempty"`
    } `json:"bottomLeft,omitempty"`
}
```

### Extracted Measurements Map

```go
type Measurement struct {
    Value     float64   `json:"value"`
    Unit      string    `json:"unit"`
    UpdatedAt time.Time `json:"updatedAt"`
}

measurements := map[string]Measurement{
    "Temp": {
        Value:     28.5,
        Unit:      "°C",
        UpdatedAt: time.Now(),
    },
    "PH": {
        Value:     7.2,
        Unit:      "",
        UpdatedAt: time.Now(),
    },
    "Redox": {
        Value:     650.0,
        Unit:      "mV",
        UpdatedAt: time.Now(),
    },
    "WaterFlow": {
        Value:     0.0,    // 0.0 = NO, 1.0 = YES
        Unit:      "NO",   // Display value
        UpdatedAt: time.Now(),
    },
}
```

## Complete Code Example

```go
func (c *AsekoClient) GetMeasurements() (map[string]Measurement, error) {
    // 1. Ensure we have a selected unit
    if c.selectedUnit == nil {
        return nil, fmt.Errorf("no unit selected")
    }

    // 2. Refresh unit data
    serialNumber := c.selectedUnit.Data.UnitBySerialNumber.SerialNumber
    if err := c.SelectUnit(serialNumber); err != nil {
        return nil, fmt.Errorf("error refreshing unit data: %w", err)
    }

    // 3. Create result map
    result := make(map[string]Measurement)

    // 4. Process primary status values
    for _, statusValue := range c.selectedUnit.Data.UnitBySerialNumber.StatusValues.Primary {
        var key string
        
        switch statusValue.Type {
        case "REDOX":
            key = "Redox"
            value, _ := strconv.ParseFloat(statusValue.Center.Value, 64)
            unit := statusValue.BottomRight
            if unit == "" {
                unit = "mV"
            }
            result[key] = Measurement{
                Value:     value,
                Unit:      unit,
                UpdatedAt: time.Now(),
            }
            
        case "PH":
            key = "PH"
            value, _ := strconv.ParseFloat(statusValue.Center.Value, 64)
            result[key] = Measurement{
                Value:     value,
                Unit:      "",
                UpdatedAt: time.Now(),
            }
            
        case "WATER_TEMPERATURE":
            key = "Temp"
            value, _ := strconv.ParseFloat(statusValue.Center.Value, 64)
            unit := statusValue.BottomRight
            if unit == "" {
                unit = "°C"
            }
            result[key] = Measurement{
                Value:     value,
                Unit:      unit,
                UpdatedAt: time.Now(),
            }
            
        case "WATER_FLOW_TO_PROBES":
            key = "WaterFlow"
            value := 0.0
            displayValue := "NO"
            if statusValue.Center.Value == "YES" {
                value = 1.0
                displayValue = "YES"
            }
            result[key] = Measurement{
                Value:     value,
                Unit:      displayValue,
                UpdatedAt: time.Now(),
            }
        }
    }

    return result, nil
}
```

## Important Notes

### 1. Value Format
- **All numeric values are stored as strings** in `center.value`
- Must parse using `strconv.ParseFloat()` or similar
- Handle parsing errors gracefully

### 2. Unit Location
- Units are typically in `bottomRight` field
- Sometimes in `bottomLeft.prefix` or `bottomLeft.suffix`
- Provide defaults if unit field is empty

### 3. Primary vs Secondary
- **Primary**: Main measurements (temp, pH, redox, flow)
- **Secondary**: Additional status info (filtration schedule, etc.)
- Focus on `primary` array for core measurements

### 4. Type Field
- The `type` field is the key to identifying measurements
- Always use exact string matching:
  - `"WATER_TEMPERATURE"` (not "temperature")
  - `"PH"` (not "ph")
  - `"REDOX"` (not "redox")
  - `"WATER_FLOW_TO_PROBES"` (not "water_flow")

### 5. Boolean/Status Values
- Water flow returns "YES" or "NO" as strings
- Convert to numeric (1.0/0.0) for processing
- Keep string value for display

### 6. Response Validation
- Always check `__typename` field
- Handle error types:
  - `"UnitNotFoundError"`
  - `"UnitAccessDeniedError"`
  - `"UnitNeverConnected"`
- Valid data type: `"Unit"`

## Error Handling

### Empty Response
```go
if response.Data.UnitBySerialNumber.SerialNumber == "" {
    // Token likely expired - reconnect
    c.Initialize()
    // Retry once
}
```

### Missing Measurements
```go
if len(response.Data.UnitBySerialNumber.StatusValues.Primary) == 0 {
    // No measurements available
    // Unit might be offline
}
```

### Parse Errors
```go
value, err := strconv.ParseFloat(statusValue.Center.Value, 64)
if err != nil {
    // Invalid number format
    // Use default value or skip
    continue
}
```

## Testing

### Sample cURL Request

```bash
curl -X POST https://graphql.acs.prod.aseko.cloud/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "X-App-Name: pool-live" \
  -H "X-App-Version: 4.2.0" \
  -H "X-Mode: production" \
  -d '{
    "operationName": "UnitDetailStatusQuery",
    "query": "query UnitDetailStatusQuery($sn: String!) { unitBySerialNumber(serialNumber: $sn) { __typename ... on Unit { statusValues { primary { type center { value } bottomRight } } } } }",
    "variables": {
      "sn": "YOUR_SERIAL_NUMBER"
    }
  }'
```

## Summary

1. **Query**: `UnitDetailStatusQuery` with unit serial number
2. **Response Path**: `data.unitBySerialNumber.statusValues.primary[]`
3. **Measurement Identification**: Use `type` field
4. **Value Extraction**: `center.value` (always string, must parse)
5. **Unit Extraction**: `bottomRight` or `bottomLeft.prefix/suffix`
6. **Processing**: Convert strings to appropriate types
7. **Result**: Map of measurement names to values with units

This approach provides reliable access to all pool measurements from the Aseko Cloud API.

