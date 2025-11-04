
# Aseko Cloud API Documentation
This document provides a comprehensive guide to integrating with the Aseko Cloud API for pool monitoring applications.
## Table of Contents
- [Overview](#overview)
- [Authentication](#authentication)
- [GraphQL API](#graphql-api)
- [WebSocket Subscriptions](#websocket-subscriptions)
- [Common Workflows](#common-workflows)
- [Data Models](#data-models)
- [Error Handling](#error-handling)
- [Best Practices](#best-practices)
## Overview
Aseko Cloud provides a modern API for accessing pool management data, including unit information, measurements, and real-time status updates. The API consists of:
- REST endpoints for authentication
- GraphQL for data queries
- WebSocket for real-time updates
Core URLs:
- Base URL: `https://aseko.cloud`
- Auth URL: `https://auth.acs.aseko.cloud/auth/login`
- GraphQL URL: `https://graphql.acs.prod.aseko.cloud/graphql`
- WebSocket URL: `wss://api.acs.prod.aseko.cloud/subscription`
## Authentication
Authentication is required for all API operations and follows these steps:
### 1. Login Request
Send a POST request to the authentication endpoint with user credentials.
```
POST https://auth.acs.aseko.cloud/auth/login
Content-Type: application/json
{
  "email": "your-email@example.com",
  "password": "your-password"
}
```
### 2. Authentication Response
The server responds with authentication tokens:
```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIsIn...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsIn...",
  "expiresIn": 3600
}
```
### 3. Using the Access Token
Include the access token in the Authorization header for all subsequent requests:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsIn...
```
### 4. Token Refresh
When the access token expires, use the refresh token to obtain a new one:
```
POST https://auth.acs.aseko.cloud/auth/refresh
Content-Type: application/json
{
  "refreshToken": "eyJhbGciOiJIUzI1NiIsIn..."
}
```
## GraphQL API
The Aseko Cloud uses GraphQL for most data operations.
### Endpoint
```
POST https://graphql.acs.prod.aseko.cloud/graphql
Content-Type: application/json
Authorization: Bearer YOUR_ACCESS_TOKEN
```
### Request Format
```json
{
  "query": "YOUR_GRAPHQL_QUERY",
  "variables": { "varName": "value" }
}
```
### Key Queries
#### Fetch User's Units
```graphql
query GetUnits {
  units {
    id
    name
    unitId
    serialNumber
    systemType
    cloudId
    mac
    status
    firmware
    updatedAt
    location {
      id
      name
      address
      timezone
    }
  }
}
```
#### Fetch Unit Details
```graphql
query GetUnitDetails($unitId: ID!) {
  unit(id: $unitId) {
    id
    name
    unitId
    serialNumber
    systemType
    cloudId
    mac
    status
    firmware
    updatedAt
    location {
      id
      name
      address
      timezone
    }
    modules {
      id
      name
      type
      serial
      firmware
    }
    measurements {
      id
      type
      value
      unit
      name
      updatedAt
    }
  }
}
```
#### Fetch Unit Measurements
```graphql
query GetUnitMeasurements($unitId: ID!) {
  unit(id: $unitId) {
    measurements {
      id
      type
      value
      unit
      name
      updatedAt
    }
  }
}
```
#### Fetch Flow Status
```graphql
query GetFlowStatus($unitId: ID!) {
  unit(id: $unitId) {
    measurements(types: ["FLOW_STATE"]) {
      id
      type
      value
      unit
      name
      updatedAt
    }
  }
}
```
## WebSocket Subscriptions
WebSockets provide real-time updates for pool data changes.
### Connection
1. Connect to the WebSocket endpoint using the access token:
   ```
   wss://api.acs.prod.aseko.cloud/subscription?token=YOUR_ACCESS_TOKEN
   ```
2. After connecting, send an initialization message:
   ```json
   {
     "type": "connection_init",
     "payload": {}
   }
   ```
3. The server should respond with:
   ```json
   {
     "type": "connection_ack"
   }
   ```
### Subscribing to Updates
To subscribe to unit updates:
```json
{
  "id": "1",
  "type": "start",
  "payload": {
    "query": "subscription UnitUpdates($cloudId: ID!) { unitUpdated(cloudId: $cloudId) { id status measurements { id type value unit updatedAt } } }",
    "variables": {
      "cloudId": "YOUR_CLOUD_ID"
    }
  }
}
```
Example CloudId: `"01HXS50KTV7NRSVNHD617J4CKB"`
### Handling WebSocket Messages
The server will send messages with the following structure for updates:
```json
{
  "id": "1",
  "type": "data",
  "payload": {
    "data": {
      "unitUpdated": {
        "id": "unit-id",
        "status": "ONLINE",
        "measurements": [
          {
            "id": "measurement-id",
            "type": "FLOW_STATE",
            "value": "1",
            "unit": "BOOLEAN",
            "updatedAt": "2023-06-01T12:00:00Z"
          }
        ]
      }
    }
  }
}
```
### Unsubscribing
To stop receiving updates:
```json
{
  "id": "1",
  "type": "stop"
}
```
### Maintaining the Connection
Send periodic ping messages to keep the connection alive:
```json
{
  "type": "ping"
}
```
## Common Workflows
### 1. Application Initialization
1. Authenticate user to get access token
2. Fetch all units using GraphQL
3. Establish WebSocket connection
4. Subscribe to unit updates
5. Display initial unit data
### 2. Monitoring Flow State
1. Fetch initial flow state using GraphQL
2. Subscribe to flow state updates via WebSocket
3. Update UI when flow state changes
4. Notify user of critical status changes
### 3. Detailed Unit Monitoring
1. Fetch detailed unit information using GraphQL
2. Subscribe to all unit measurements
3. Create real-time dashboard with relevant metrics
4. Implement notifications for abnormal values
## Data Models
### Unit
```json
{
  "id": "string",
  "name": "string",
  "unitId": "string",
  "serialNumber": "string",
  "systemType": "string",
  "cloudId": "string",
  "mac": "string",
  "status": "ONLINE|OFFLINE|DISCONNECTED",
  "firmware": "string",
  "updatedAt": "ISO8601 timestamp"
}
```
### Location
```json
{
  "id": "string",
  "name": "string",
  "address": "string",
  "timezone": "string"
}
```
### Measurement
```json
{
  "id": "string",
  "type": "FLOW_STATE|PH|REDOX|TEMPERATURE|...",
  "value": "string",
  "unit": "BOOLEAN|PH|MV|CELSIUS|...",
  "name": "string",
  "updatedAt": "ISO8601 timestamp"
}
```
### Module
```json
{
  "id": "string",
  "name": "string",
  "type": "string",
  "serial": "string",
  "firmware": "string"
}
```
## Error Handling
### GraphQL Errors
GraphQL errors are returned in the response:
```json
{
  "errors": [
    {
      "message": "Error message",
      "path": ["query", "field"],
      "extensions": {
        "code": "ERROR_CODE"
      }
    }
  ]
}
```
Common error codes:
- `UNAUTHENTICATED`: Invalid or expired token
- `FORBIDDEN`: Insufficient permissions
- `NOT_FOUND`: Requested resource not found
- `BAD_REQUEST`: Invalid request format
### WebSocket Errors
WebSocket errors are sent as messages:
```json
{
  "type": "error",
  "id": "subscription-id",
  "payload": {
    "message": "Error message"
  }
}
```
### Authentication Errors
Authentication errors are returned as HTTP response:
```json
{
  "error": "invalid_grant",
  "error_description": "Bad credentials"
}
```
## Best Practices
1. **Token Management**
   - Store tokens securely
   - Implement token refresh before expiration
   - Handle authentication failures gracefully
2. **Data Polling vs. WebSockets**
   - Use WebSockets for real-time updates
   - Fall back to polling if WebSockets are unavailable
   - Implement reconnection logic with exponential backoff
3. **Error Handling**
   - Implement comprehensive error handling
   - Provide meaningful feedback to users
   - Log errors for debugging
4. **Performance Optimization**
   - Request only needed fields in GraphQL queries
   - Implement caching for frequently accessed data
   - Batch related GraphQL queries
5. **User Experience**
   - Show connection status clearly
   - Provide visual feedback for data freshness
   - Implement notifications for critical events
6. **Security**
   - Never store credentials in client-side code
   - Implement secure token storage
   - Use HTTPS for all requests
   - Validate all user inputs
7. **Testing**
   - Test with intermittent connectivity
   - Verify reconnection behaviors
   - Test token refresh workflows
