package com.example.asekoflowmonitor.service;

import com.example.asekoflowmonitor.config.AsekoConfig;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.client.methods.HttpOptions;
import org.apache.http.entity.ContentType;
import org.apache.http.entity.StringEntity;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.util.EntityUtils;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

import javax.annotation.PostConstruct;
import java.io.IOException;
import java.nio.charset.StandardCharsets;

@Service
public class UnitService {

    private final AsekoConfig asekoConfig;
    private final AuthService authService;
    private final ObjectMapper objectMapper;
    private final SimpMessagingTemplate messagingTemplate;
    private JsonNode unitListData;
    private JsonNode selectedUnit;

    @Autowired
    public UnitService(AsekoConfig asekoConfig, AuthService authService, SimpMessagingTemplate messagingTemplate) {
        this.asekoConfig = asekoConfig;
        this.authService = authService;
        this.objectMapper = new ObjectMapper();
        this.messagingTemplate = messagingTemplate;
    }

    @PostConstruct
    public void init() {
        new Thread(() -> {
            try {
                Thread.sleep(8000); // Wait for authentication to complete
                fetchUnitList();
                
                // Select the first unit after fetching the list
                if (unitListData != null && unitListData.path("units").isArray() 
                        && unitListData.path("units").size() > 0) {
                    selectUnit(unitListData.path("units").get(0));
                }
            } catch (Exception e) {
                System.err.println("Error initializing UnitService: " + e.getMessage());
            }
        }).start();
    }

    @Scheduled(fixedRate = 300000) // Refresh unit list every 5 minutes
    public void refreshUnitList() {
        try {
            fetchUnitList();
        } catch (Exception e) {
            System.err.println("Error refreshing unit list: " + e.getMessage());
        }
    }

    public JsonNode getUnitList() {
        return unitListData;
    }

    public void fetchUnitList() throws IOException {
        System.out.println("\n===== FETCHING UNIT LIST =====");
        
        try {
            // Make sure we have a valid auth token
            String token = authService.getAuthToken();
            if (token == null || token.isEmpty()) {
                throw new IOException("Authentication required");
            }
            
            // Create the HTTP client
            try (CloseableHttpClient httpClient = HttpClients.createDefault()) {
                HttpPost httpPost = new HttpPost("https://graphql.acs.prod.aseko.cloud/graphql");
                
                // Set headers
                httpPost.setHeader("Accept", "*/*");
                httpPost.setHeader("Accept-Language", "en");
                httpPost.setHeader("Authorization", "Bearer " + token);
                httpPost.setHeader("Connection", "keep-alive");
                httpPost.setHeader("Content-Type", "application/json");
                httpPost.setHeader("Origin", "https://aseko.cloud");
                httpPost.setHeader("Referer", "https://aseko.cloud/");
                httpPost.setHeader("Sec-Fetch-Dest", "empty");
                httpPost.setHeader("Sec-Fetch-Mode", "cors");
                httpPost.setHeader("Sec-Fetch-Site", "same-site");
                httpPost.setHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36");
                httpPost.setHeader("X-App-Name", "pool-live");
                httpPost.setHeader("X-App-Version", "4.2.0");
                httpPost.setHeader("X-Mode", "production");
                
                // Create GraphQL query for UnitList
                ObjectNode queryBody = objectMapper.createObjectNode();
                queryBody.put("operationName", "UnitList");
                
                ObjectNode variables = objectMapper.createObjectNode();
                variables.putNull("after");
                variables.put("first", 15);
                variables.put("search", "");
                queryBody.set("variables", variables);
                
                // Use the exact GraphQL query from the curl command
                String query = "fragment UnitFragment on Unit {\n" +
                               "  __typename\n" +
                               "  serialNumber\n" +
                               "  name\n" +
                               "  note\n" +
                               "  brandName {\n" +
                               "    id\n" +
                               "    primary\n" +
                               "    secondary\n" +
                               "    __typename\n" +
                               "  }\n" +
                               "  position\n" +
                               "  statusMessages {\n" +
                               "    __typename\n" +
                               "    type\n" +
                               "    severity\n" +
                               "    message\n" +
                               "  }\n" +
                               "  consumables {\n" +
                               "    __typename\n" +
                               "    ... on LiquidConsumable {\n" +
                               "      canister {\n" +
                               "        __typename\n" +
                               "        id\n" +
                               "        hasWarning\n" +
                               "      }\n" +
                               "      tube {\n" +
                               "        __typename\n" +
                               "        id\n" +
                               "        hasWarning\n" +
                               "      }\n" +
                               "      __typename\n" +
                               "    }\n" +
                               "    ... on ElectrolyzerConsumable {\n" +
                               "      electrode {\n" +
                               "        __typename\n" +
                               "        hasWarning\n" +
                               "      }\n" +
                               "      __typename\n" +
                               "    }\n" +
                               "  }\n" +
                               "  online\n" +
                               "  offlineFor\n" +
                               "  hasWarning\n" +
                               "  notificationConfiguration {\n" +
                               "    __typename\n" +
                               "    id\n" +
                               "    hasWarning\n" +
                               "  }\n" +
                               "  unitModel {\n" +
                               "    __typename\n" +
                               "    id\n" +
                               "    tabs {\n" +
                               "      hideNotifications\n" +
                               "      hideConsumables\n" +
                               "      __typename\n" +
                               "    }\n" +
                               "  }\n" +
                               "}\n\n" +
                               "fragment UnitNeverConnectedFragment on UnitNeverConnected {\n" +
                               "  __typename\n" +
                               "  serialNumber\n" +
                               "  name\n" +
                               "  note\n" +
                               "  position\n" +
                               "  statusMessages {\n" +
                               "    __typename\n" +
                               "    severity\n" +
                               "    type\n" +
                               "    message\n" +
                               "    detail\n" +
                               "  }\n" +
                               "}\n\n" +
                               "query UnitList($after: String, $first: Int, $search: String) {\n" +
                               "  units(after: $after, first: $first, searchQuery: $search) {\n" +
                               "    cursor\n" +
                               "    units {\n" +
                               "      ...UnitFragment\n" +
                               "      ...UnitNeverConnectedFragment\n" +
                               "      __typename\n" +
                               "    }\n" +
                               "    __typename\n" +
                               "  }\n" +
                               "}";
                queryBody.put("query", query);
                
                // Create the request entity
                StringEntity entity = new StringEntity(objectMapper.writeValueAsString(queryBody), ContentType.APPLICATION_JSON);
                httpPost.setEntity(entity);
                
                // Execute the request
                try (CloseableHttpResponse response = httpClient.execute(httpPost)) {
                    int statusCode = response.getStatusLine().getStatusCode();
                    String responseBody = EntityUtils.toString(response.getEntity(), StandardCharsets.UTF_8);
                    
                    System.out.println("Unit list query response status: " + statusCode);
                    
                    if (statusCode == 200) {
                        JsonNode jsonResponse = objectMapper.readTree(responseBody);
                        
                        // Check for errors
                        JsonNode errors = jsonResponse.path("errors");
                        if (errors.isArray() && errors.size() > 0) {
                            System.err.println("GraphQL errors: " + errors);
                            return;
                        }
                        
                        // Store the unit list data
                        this.unitListData = jsonResponse.path("data").path("units");
                        
                        // Count the units
                        int unitCount = 0;
                        if (this.unitListData.path("units").isArray()) {
                            unitCount = this.unitListData.path("units").size();
                        }
                        
                        System.out.println("Successfully fetched " + unitCount + " units");
                        
                        // Send to connected clients
                        messagingTemplate.convertAndSend("/topic/unitList", this.unitListData);
                    } else {
                        System.err.println("Unit list query failed, status: " + statusCode);
                        System.err.println("Response: " + responseBody);
                    }
                }
            }
        } catch (Exception e) {
            System.err.println("Error fetching unit list: " + e.getMessage());
            e.printStackTrace();
            throw new IOException("Failed to fetch unit list: " + e.getMessage(), e);
        }
        
        System.out.println("===== UNIT LIST FETCH COMPLETE =====\n");
    }

    public JsonNode getSelectedUnit() {
        return selectedUnit;
    }

    public void selectUnit(JsonNode unit) throws IOException {
        System.out.println("\n===== SELECTING UNIT =====");
        
        try {
            if (unit == null) {
                throw new IOException("Cannot select null unit");
            }
            
            String serialNumber = unit.path("serialNumber").asText();
            String name = unit.path("name").asText("Unnamed Unit");
            
            System.out.println("Selecting unit: " + name + " (Serial: " + serialNumber + ")");
            
            // Make sure we have a valid auth token
            String token = authService.getAuthToken();
            if (token == null || token.isEmpty()) {
                throw new IOException("Authentication required");
            }
            
            // Store the selected unit
            this.selectedUnit = unit;
            
            // Send the selected unit to clients
            messagingTemplate.convertAndSend("/topic/selectedUnit", unit);
            
            // Fetch detailed information for this unit
            if (serialNumber != null && !serialNumber.isEmpty()) {
                JsonNode unitDetail = fetchUnitDetail(serialNumber);
                if (unitDetail != null) {
                    // Update the selected unit with more detailed information
                    this.selectedUnit = unitDetail;
                    messagingTemplate.convertAndSend("/topic/selectedUnit", unitDetail);
                }
            }
            
            System.out.println("Unit selected successfully: " + name);
        } catch (Exception e) {
            System.err.println("Error selecting unit: " + e.getMessage());
            e.printStackTrace();
            throw new IOException("Failed to select unit: " + e.getMessage(), e);
        }
        
        System.out.println("===== UNIT SELECTION COMPLETE =====\n");
    }
    
    /**
     * When fetching the unit list, this method extracts the unit ID from the configuration
     * and selects that unit if it exists in the list.
     */
    private void selectConfiguredUnit() {
        if (unitListData == null || !unitListData.path("units").isArray()) {
            return;
        }
        
        String configUnitId = asekoConfig.getUnitId();
        if (configUnitId == null || configUnitId.isEmpty()) {
            System.out.println("No unit ID configured, will select first unit by default");
            return;
        }
        
        System.out.println("Looking for configured unit ID: " + configUnitId);
        
        // Look for the configured unit in the list
        JsonNode units = unitListData.path("units");
        for (JsonNode unit : units) {
            String id = unit.path("id").asText();
            if (configUnitId.equals(id)) {
                try {
                    selectUnit(unit);
                    return;
                } catch (IOException e) {
                    System.err.println("Failed to select configured unit: " + e.getMessage());
                }
            }
        }
        
        // If we didn't find the configured unit, select the first one
        try {
            if (units.size() > 0) {
                System.out.println("Configured unit not found, selecting first unit instead");
                selectUnit(units.get(0));
            }
        } catch (IOException e) {
            System.err.println("Failed to select first unit: " + e.getMessage());
        }
    }

    public JsonNode fetchUnitDetail(String serialNumber) throws IOException {
        System.out.println("\n===== FETCHING UNIT DETAILS FOR " + serialNumber + " =====");
        
        try {
            // Make sure we have a valid auth token by refreshing the login
            authService.login();
            String token = authService.getAuthToken();
            if (token == null || token.isEmpty()) {
                throw new IOException("Authentication required");
            }
            
            // Create the HTTP client
            try (CloseableHttpClient httpClient = HttpClients.createDefault()) {
                HttpPost httpPost = new HttpPost("https://graphql.acs.prod.aseko.cloud/graphql");
                
                // Set headers
                httpPost.setHeader("Connection", "keep-alive");
                httpPost.setHeader("Origin", "https://aseko.cloud");
                httpPost.setHeader("Referer", "https://aseko.cloud/");
                httpPost.setHeader("Sec-Fetch-Dest", "empty");
                httpPost.setHeader("Sec-Fetch-Mode", "cors");
                httpPost.setHeader("Sec-Fetch-Site", "same-site");
                httpPost.setHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36");
                httpPost.setHeader("accept", "*/*");
                httpPost.setHeader("accept-language", "en");
                httpPost.setHeader("authorization", "Bearer " + token);
                httpPost.setHeader("content-type", "application/json");
                httpPost.setHeader("sec-ch-ua", "\"Not(A:Brand\";v=\"99\", \"Google Chrome\";v=\"133\", \"Chromium\";v=\"133\"");
                httpPost.setHeader("sec-ch-ua-mobile", "?0");
                httpPost.setHeader("sec-ch-ua-platform", "\"macOS\"");
                httpPost.setHeader("x-app-name", "pool-live");
                httpPost.setHeader("x-app-version", "4.2.0");
                httpPost.setHeader("x-mode", "production");
                
                // Create the request payload PROPERLY with no escaping issues
                ObjectNode requestPayload = objectMapper.createObjectNode();
                requestPayload.put("operationName", "UnitDetailStatusQuery");
                
                // Set variables
                ObjectNode variables = objectMapper.createObjectNode();
                variables.put("sn", serialNumber);
                requestPayload.set("variables", variables);
                
                // Define the GraphQL query with proper Java escaping
                // Let Jackson handle the JSON escaping
                String query = "fragment StatusValueFragment on StatusValue {\n" +
                        "  __typename\n" +
                        "  id\n" +
                        "  type\n" +
                        "  backgroundColor\n" +
                        "  textColor\n" +
                        "  topLeft\n" +
                        "  topRight\n" +
                        "  center {\n" +
                        "    __typename\n" +
                        "    ... on StringValue {\n" +
                        "      value\n" +
                        "      iconName\n" +
                        "      __typename\n" +
                        "    }\n" +
                        "    ... on UpcomingFiltrationPeriodValue {\n" +
                        "      __typename\n" +
                        "      configuration {\n" +
                        "        __typename\n" +
                        "        name\n" +
                        "        speed\n" +
                        "        start\n" +
                        "        end\n" +
                        "        overrideIntervalText\n" +
                        "        poolFlow\n" +
                        "      }\n" +
                        "      isNext\n" +
                        "    }\n" +
                        "  }\n" +
                        "  bottomRight\n" +
                        "  bottomLeft {\n" +
                        "    __typename\n" +
                        "    prefix\n" +
                        "    suffix\n" +
                        "    style\n" +
                        "  }\n" +
                        "}\n" +
                        "\n" +
                        "fragment BackwashStatusFragment on BackwashStatus {\n" +
                        "  __typename\n" +
                        "  id\n" +
                        "  running\n" +
                        "  duration\n" +
                        "  elapsed\n" +
                        "  configuration {\n" +
                        "    __typename\n" +
                        "    oncePerXDays\n" +
                        "    start\n" +
                        "    takes\n" +
                        "  }\n" +
                        "}\n" +
                        "\n" +
                        "fragment StatusMessageFragment on StatusMessage {\n" +
                        "  __typename\n" +
                        "  type\n" +
                        "  severity\n" +
                        "  message\n" +
                        "  detail\n" +
                        "}\n" +
                        "\n" +
                        "query UnitDetailStatusQuery($sn: String!) {\n" +
                        "  unitBySerialNumber(serialNumber: $sn) {\n" +
                        "    __typename\n" +
                        "    ... on UnitNotFoundError {\n" +
                        "      serialNumber\n" +
                        "      __typename\n" +
                        "    }\n" +
                        "    ... on UnitAccessDeniedError {\n" +
                        "      serialNumber\n" +
                        "      __typename\n" +
                        "    }\n" +
                        "    ... on UnitNeverConnected {\n" +
                        "      serialNumber\n" +
                        "      name\n" +
                        "      note\n" +
                        "      statusMessages {\n" +
                        "        __typename\n" +
                        "        type\n" +
                        "        message\n" +
                        "        severity\n" +
                        "        detail\n" +
                        "      }\n" +
                        "      __typename\n" +
                        "    }\n" +
                        "    ... on Unit {\n" +
                        "      serialNumber\n" +
                        "      name\n" +
                        "      note\n" +
                        "      statusMessages {\n" +
                        "        __typename\n" +
                        "        type\n" +
                        "        message\n" +
                        "        severity\n" +
                        "        detail\n" +
                        "      }\n" +
                        "      offlineFor\n" +
                        "      statusValues {\n" +
                        "        __typename\n" +
                        "        primary {\n" +
                        "          ...StatusValueFragment\n" +
                        "          __typename\n" +
                        "        }\n" +
                        "        secondary {\n" +
                        "          ...StatusValueFragment\n" +
                        "          __typename\n" +
                        "        }\n" +
                        "      }\n" +
                        "      statusMessages {\n" +
                        "        ...StatusMessageFragment\n" +
                        "        __typename\n" +
                        "      }\n" +
                        "      backwash {\n" +
                        "        ...BackwashStatusFragment\n" +
                        "        __typename\n" +
                        "      }\n" +
                        "      waterFilling {\n" +
                        "        __typename\n" +
                        "        id\n" +
                        "        waterLevel\n" +
                        "        totalTime\n" +
                        "        totalLiters\n" +
                        "        totalTimeFromLastReset\n" +
                        "        totalLitersFromLastReset\n" +
                        "        lastReset\n" +
                        "        litersPerMinute\n" +
                        "        configuration {\n" +
                        "          __typename\n" +
                        "          levelHigh\n" +
                        "          levelLow\n" +
                        "          levelMax\n" +
                        "          levelMin\n" +
                        "          maxFillingTime\n" +
                        "          enabled\n" +
                        "        }\n" +
                        "      }\n" +
                        "      __typename\n" +
                        "    }\n" +
                        "  }\n" +
                        "}";
                
                requestPayload.put("query", query);
                
                // Convert to JSON string with proper escaping handled by Jackson
                String jsonPayload = objectMapper.writeValueAsString(requestPayload);
                System.out.println("Sending properly formatted request");
                
                // Create the request entity
                StringEntity entity = new StringEntity(jsonPayload, ContentType.APPLICATION_JSON);
                httpPost.setEntity(entity);
                
                // Execute the request
                try (CloseableHttpResponse response = httpClient.execute(httpPost)) {
                    int statusCode = response.getStatusLine().getStatusCode();
                    String responseBody = EntityUtils.toString(response.getEntity(), StandardCharsets.UTF_8);
                    
                    System.out.println("Unit detail query response status: " + statusCode);
                    
                    if (statusCode == 200) {
                        JsonNode jsonResponse = objectMapper.readTree(responseBody);
                        
                        // Check for errors
                        JsonNode errors = jsonResponse.path("errors");
                        if (errors.isArray() && errors.size() > 0) {
                            System.err.println("GraphQL errors: " + errors);
                            return null;
                        }
                        
                        // Extract the unit detail data
                        JsonNode unitDetail = jsonResponse.path("data").path("unitBySerialNumber");
                        
                        if (unitDetail != null) {
                            String typename = unitDetail.path("__typename").asText();
                            if ("UnitNotFoundError".equals(typename) || "UnitAccessDeniedError".equals(typename)) {
                                System.err.println("Error fetching unit: " + typename);
                                return null;
                            }
                            
                            System.out.println("Successfully fetched details for unit: " + serialNumber);
                            
                            // Send to connected clients
                            messagingTemplate.convertAndSend("/topic/unitDetail", unitDetail);
                            
                            return unitDetail;
                        }
                    } else {
                        System.err.println("Unit detail query failed, status: " + statusCode);
                        System.err.println("Response: " + responseBody);
                    }
                }
            }
        } catch (Exception e) {
            System.err.println("Error fetching unit details: " + e.getMessage());
            e.printStackTrace();
            throw new IOException("Failed to fetch unit details: " + e.getMessage(), e);
        }
        
        System.out.println("===== UNIT DETAIL FETCH COMPLETE =====\n");
        return null;
    }

    /**
     * Alternative approach for fetching unit details with explicit OPTIONS preflight
     */
    private JsonNode tryAlternativeUnitDetailFetch(String serialNumber, String token) throws IOException {
        System.out.println("\n===== TRYING ALTERNATIVE UNIT DETAILS FETCH =====");
        try {
            CloseableHttpClient httpClient = HttpClients.createDefault();
            
            // First send OPTIONS request (preflight)
            HttpOptions optionsRequest = new HttpOptions("https://graphql.acs.prod.aseko.cloud/graphql");
            optionsRequest.setHeader("Accept", "*/*");
            optionsRequest.setHeader("Accept-Language", "en-GB,en;q=0.9,en-US;q=0.8");
            optionsRequest.setHeader("Access-Control-Request-Headers", "authorization,content-type,x-app-name,x-app-version,x-mode");
            optionsRequest.setHeader("Access-Control-Request-Method", "POST");
            optionsRequest.setHeader("Connection", "keep-alive");
            optionsRequest.setHeader("Origin", "https://aseko.cloud");
            optionsRequest.setHeader("Referer", "https://aseko.cloud/");
            optionsRequest.setHeader("Sec-Fetch-Dest", "empty");
            optionsRequest.setHeader("Sec-Fetch-Mode", "cors");
            optionsRequest.setHeader("Sec-Fetch-Site", "same-site");
            optionsRequest.setHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36");
            
            try (CloseableHttpResponse optionsResponse = httpClient.execute(optionsRequest)) {
                int optionsStatusCode = optionsResponse.getStatusLine().getStatusCode();
                System.out.println("OPTIONS preflight status: " + optionsStatusCode);
                
                // Now send the actual POST request
                HttpPost httpPost = new HttpPost("https://graphql.acs.prod.aseko.cloud/graphql");
                httpPost.setHeader("Connection", "keep-alive");
                httpPost.setHeader("Origin", "https://aseko.cloud");
                httpPost.setHeader("Referer", "https://aseko.cloud/");
                httpPost.setHeader("Sec-Fetch-Dest", "empty");
                httpPost.setHeader("Sec-Fetch-Mode", "cors");
                httpPost.setHeader("Sec-Fetch-Site", "same-site");
                httpPost.setHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36");
                httpPost.setHeader("accept", "*/*");
                httpPost.setHeader("accept-language", "en");
                httpPost.setHeader("authorization", "Bearer " + token);
                httpPost.setHeader("content-type", "application/json");
                httpPost.setHeader("sec-ch-ua", "\"Not(A:Brand\";v=\"99\", \"Google Chrome\";v=\"133\", \"Chromium\";v=\"133\"");
                httpPost.setHeader("sec-ch-ua-mobile", "?0");
                httpPost.setHeader("sec-ch-ua-platform", "\"macOS\"");
                httpPost.setHeader("x-app-name", "pool-live");
                httpPost.setHeader("x-app-version", "4.2.0");
                httpPost.setHeader("x-mode", "production");
                
                // Create exact same request body
                ObjectNode queryBody = objectMapper.createObjectNode();
                queryBody.put("operationName", "UnitDetail");
                ObjectNode variables = objectMapper.createObjectNode();
                variables.put("sn", serialNumber);
                queryBody.set("variables", variables);
                queryBody.put("query", "query UnitDetail($sn: String\\u0021) {\\n  unitBySerialNumber(serialNumber: $sn) {\\n    __typename\\n    ... on UnitNotFoundError {\\n      serialNumber\\n      __typename\\n    }\\n    ... on UnitAccessDeniedError {\\n      serialNumber\\n      __typename\\n    }\\n    ... on UnitNeverConnected {\\n      serialNumber\\n      name\\n      note\\n      statusMessages {\\n        __typename\\n        severity\\n        type\\n        message\\n        detail\\n      }\\n      __typename\\n    }\\n    ... on Unit {\\n      serialNumber\\n      name\\n      note\\n      brandName {\\n        __typename\\n        id\\n        primary\\n        secondary\\n      }\\n      statusMessages {\\n        __typename\\n        severity\\n        type\\n        message\\n        detail\\n      }\\n      heating {\\n        __typename\\n        lastReset\\n      }\\n      waterFilling {\\n        __typename\\n        id\\n        waterLevel\\n        lastReset\\n      }\\n      consumables {\\n        __typename\\n        ... on LiquidConsumable {\\n          type\\n          canister {\\n            __typename\\n            id\\n            hasWarning\\n          }\\n          tube {\\n            __typename\\n            id\\n            hasWarning\\n          }\\n          __typename\\n        }\\n        ... on ElectrolyzerConsumable {\\n          type\\n          electrode {\\n            __typename\\n            hasWarning\\n          }\\n          __typename\\n        }\\n      }\\n      notificationConfiguration {\\n        __typename\\n        id\\n        type\\n        name\\n        enabled\\n        lowWarningLevel\\n        highWarningLevel\\n        color\\n        currentValue\\n        suffix\\n        hasWarning\\n        possibleWarningLevels\\n      }\\n      unitModel {\\n        __typename\\n        id\\n        tabs {\\n          hideNotifications\\n          hideConsumables\\n          hideProtocolExport\\n          __typename\\n        }\\n      }\\n      __typename\\n    }\\n  }\\n}");
                
                String jsonPayload = objectMapper.writeValueAsString(queryBody);
                StringEntity entity = new StringEntity(jsonPayload, ContentType.APPLICATION_JSON);
                httpPost.setEntity(entity);
                
                try (CloseableHttpResponse response = httpClient.execute(httpPost)) {
                    int statusCode = response.getStatusLine().getStatusCode();
                    String responseBody = EntityUtils.toString(response.getEntity(), StandardCharsets.UTF_8);
                    
                    System.out.println("Alternative unit detail request status: " + statusCode);
                    
                    if (statusCode == 200) {
                        JsonNode jsonResponse = objectMapper.readTree(responseBody);
                        JsonNode unitDetail = jsonResponse.path("data").path("unitBySerialNumber");
                        
                        if (unitDetail != null) {
                            String typename = unitDetail.path("__typename").asText();
                            if ("UnitNotFoundError".equals(typename) || "UnitAccessDeniedError".equals(typename)) {
                                System.err.println("Error fetching unit: " + typename);
                                return null;
                            }
                            
                            System.out.println("Successfully fetched unit details using alternative method");
                            messagingTemplate.convertAndSend("/topic/unitDetail", unitDetail);
                            return unitDetail;
                        }
                    } else {
                        System.err.println("Alternative unit detail request failed: " + statusCode);
                        System.err.println("Response: " + responseBody);
                    }
                }
            }
        } catch (Exception e) {
            System.err.println("Error in alternative unit fetch: " + e.getMessage());
        }
        
        System.out.println("===== ALTERNATIVE FETCH COMPLETED =====\n");
        return null;
    }
} 