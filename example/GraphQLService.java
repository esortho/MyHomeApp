package com.example.asekoflowmonitor.service;

import com.example.asekoflowmonitor.config.AsekoConfig;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpGet;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.entity.ContentType;
import org.apache.http.entity.StringEntity;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.util.EntityUtils;
import org.java_websocket.client.WebSocketClient;
import org.java_websocket.handshake.ServerHandshake;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.stereotype.Service;

import javax.annotation.PostConstruct;
import java.io.IOException;
import java.net.URI;
import java.net.URISyntaxException;
import java.nio.ByteBuffer;
import java.nio.charset.StandardCharsets;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.function.Consumer;

@Service
public class GraphQLService {
    
    private final AsekoConfig asekoConfig;
    private final AuthService authService;
    private final ObjectMapper objectMapper;
    private final SimpMessagingTemplate messagingTemplate;
    
    // GraphQL WebSocket settings
    private static final String GRAPHQL_WS_URL = "wss://graphql.acs.prod.aseko.cloud/graphql";
    private static final String CLOUD_ID = "01HXS50KTV7NRSVNHD617J4CKB";
    private static final String UNIT_ID = "01HXS5GVHJEHGNVJZ2YKQSBVWM"; // Replace with your unit ID
    
    private WebSocketClient client;
    private AtomicBoolean flowStatus = new AtomicBoolean(false);
    private Consumer<Boolean> flowStatusConsumer;
    
    @Autowired
    public GraphQLService(AsekoConfig asekoConfig, AuthService authService, SimpMessagingTemplate messagingTemplate) {
        this.asekoConfig = asekoConfig;
        this.authService = authService;
        this.objectMapper = new ObjectMapper();
        this.messagingTemplate = messagingTemplate;
    }
    
    @PostConstruct
    public void initialize() {
        new Thread(() -> {
            try {
                Thread.sleep(5000); // Give time for authentication
                connectWebSocket((status) -> {
                    this.flowStatus.set(status);
                    messagingTemplate.convertAndSend("/topic/status", status);
                });
            } catch (Exception e) {
                System.err.println("Error initializing WebSocket: " + e.getMessage());
                e.printStackTrace();
            }
        }).start();
    }
    
    public boolean getFlowStatus() throws IOException {
        try {
            // Get unit measurements using GraphQL
            JsonNode unitData = getUnitData();
            
            // Extract waterflow measurement
            if (unitData != null) {
                JsonNode measurements = unitData.path("measurements");
                if (!measurements.isMissingNode()) {
                    JsonNode waterflow = measurements.path("waterflow");
                    if (!waterflow.isMissingNode()) {
                        double flowValue = waterflow.asDouble();
                        System.out.println("Current waterflow: " + flowValue);
                        return flowValue > 0; // Flow is considered active if > 0
                    }
                }
            }
            
            return false; // Default to no flow if data is missing
        } catch (Exception e) {
            System.err.println("Error getting flow status: " + e.getMessage());
            throw new IOException("Failed to get flow status", e);
        }
    }
    
    private JsonNode getUnitData() throws IOException {
        try {
            // Make sure we're authenticated first
            String token = authService.getAuthToken();
            if (token == null || token.isEmpty()) {
                throw new IOException("Authentication required");
            }
            
            String graphqlEndpoint = asekoConfig.getApiUrl() + "/graphql";
            System.out.println("GraphQL Endpoint: " + graphqlEndpoint);
            
            HttpPost httpPost = new HttpPost(graphqlEndpoint);
            
            // Set headers
            httpPost.setHeader("Accept", "application/json");
            httpPost.setHeader("Content-Type", "application/json");
            httpPost.setHeader("Authorization", "Bearer " + token);
            httpPost.setHeader("Origin", "https://aseko.cloud");
            httpPost.setHeader("Referer", "https://aseko.cloud/");
            httpPost.setHeader("X-App-Name", "pool-live");
            httpPost.setHeader("X-App-Version", "4.2.0");
            httpPost.setHeader("X-Mode", "production");
            
            // Create GraphQL query
            ObjectNode queryNode = objectMapper.createObjectNode();
            queryNode.put("query", 
                "query UnitQuery($unitId: String!) {" +
                "  unit(id: $unitId) {" +
                "    id" +
                "    measurements {" +
                "      ph" +
                "      rx" +
                "      temperature" +
                "      waterflow" +
                "    }" +
                "  }" +
                "}");
            
            // Add variables
            ObjectNode variablesNode = objectMapper.createObjectNode();
            variablesNode.put("unitId", asekoConfig.getUnitId());
            queryNode.set("variables", variablesNode);
            
            // Set request body
            StringEntity entity = new StringEntity(objectMapper.writeValueAsString(queryNode), ContentType.APPLICATION_JSON);
            httpPost.setEntity(entity);
            
            // Execute request
            try (CloseableHttpClient httpClient = HttpClients.createDefault();
                 CloseableHttpResponse response = httpClient.execute(httpPost)) {
                
                int statusCode = response.getStatusLine().getStatusCode();
                String responseBody = EntityUtils.toString(response.getEntity(), StandardCharsets.UTF_8);
                
                System.out.println("GraphQL response status: " + statusCode);
                
                if (statusCode == 200) {
                    JsonNode jsonResponse = objectMapper.readTree(responseBody);
                    
                    // Check for errors
                    JsonNode errors = jsonResponse.path("errors");
                    if (errors.isArray() && errors.size() > 0) {
                        System.err.println("GraphQL errors: " + errors);
                        return null;
                    }
                    
                    // Extract data
                    JsonNode data = jsonResponse.path("data");
                    if (!data.isMissingNode()) {
                        JsonNode unit = data.path("unit");
                        if (!unit.isMissingNode()) {
                            return unit;
                        }
                    }
                    
                    System.err.println("No unit data in response: " + responseBody);
                    return null;
                } else {
                    System.err.println("GraphQL request failed, status: " + statusCode);
                    System.err.println("Response: " + responseBody);
                    return null;
                }
            } 
        } catch (Exception e) {
            System.err.println("Error in GraphQL request: " + e.getMessage());
            throw new IOException("Failed to get unit data: " + e.getMessage(), e);
        }
    }
    
    public void connectWebSocket(Consumer<Boolean> consumer) {
        try {
            String token = authService.getAuthToken();
            this.flowStatusConsumer = consumer;
            
            System.out.println("Connecting to GraphQL WebSocket: " + GRAPHQL_WS_URL);
            System.out.println("Using auth token starting with: " + 
                (token != null && token.length() > 10 ? token.substring(0, 10) + "..." : "null"));
            
            // Create WebSocket client
            Map<String, String> headers = new HashMap<>();
            headers.put("Authorization", "Bearer " + token);
            headers.put("Accept", "*/*");
            headers.put("Accept-Language", "en-GB,en;q=0.9");
            headers.put("Origin", "https://aseko.cloud");
            headers.put("Referer", "https://aseko.cloud/");
            headers.put("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36");
            headers.put("Sec-Fetch-Dest", "empty");
            headers.put("Sec-Fetch-Mode", "cors");
            headers.put("Sec-Fetch-Site", "cross-site");
            
            client = new WebSocketClient(new URI(GRAPHQL_WS_URL), headers) {
                @Override
                public void onOpen(ServerHandshake handshakedata) {
                    System.out.println("WebSocket Connected!");
                    
                    // Send connection init message
                    try {
                        Thread.sleep(1000);
                        sendConnectionInit(token);
                    } catch (Exception e) {
                        System.err.println("Error sending connection init: " + e.getMessage());
                    }
                }
                
                @Override
                public void onMessage(String message) {
                    try {
                        System.out.println("Received message: " + message);
                        JsonNode node = objectMapper.readTree(message);
                        String messageType = node.path("type").asText();
                        
                        if ("connection_ack".equals(messageType)) {
                            System.out.println("Connection acknowledged, subscribing to updates...");
                            sendSubscription();
                        } else if ("data".equals(messageType)) {
                            handleDataMessage(node);
                        } else if ("error".equals(messageType)) {
                            System.err.println("WebSocket error: " + node.path("payload").toString());
                        }
                    } catch (Exception e) {
                        System.err.println("Error processing message: " + e.getMessage());
                    }
                }
                
                @Override
                public void onClose(int code, String reason, boolean remote) {
                    System.out.println("WebSocket connection closed: " + code + " " + reason);
                    
                    // Reconnect after a delay
                    new Thread(() -> {
                        try {
                            Thread.sleep(5000);
                            System.out.println("Attempting to reconnect WebSocket...");
                            connectWebSocket(flowStatusConsumer);
                        } catch (Exception e) {
                            System.err.println("Error reconnecting: " + e.getMessage());
                        }
                    }).start();
                }
                
                @Override
                public void onError(Exception ex) {
                    System.err.println("WebSocket error: " + ex.getMessage());
                    ex.printStackTrace();
                }
                
                @Override
                public void onMessage(ByteBuffer bytes) {
                    System.out.println("Received binary message");
                }
            };
            
            client.connect();
            
        } catch (URISyntaxException e) {
            System.err.println("Error connecting to WebSocket: " + e.getMessage());
            e.printStackTrace();
        }
    }
    
    private void sendConnectionInit(String token) {
        try {
            ObjectNode initMessage = objectMapper.createObjectNode();
            initMessage.put("type", "connection_init");
            
            ObjectNode payload = objectMapper.createObjectNode();
            ObjectNode headers = objectMapper.createObjectNode();
            headers.put("Authorization", "Bearer " + token);
            headers.put("X-App-Name", "pool-live");
            headers.put("X-App-Version", "4.1.0");
            headers.put("X-Mode", "production");
            headers.put("X-Cloud", asekoConfig.getCloudId());
            
            payload.set("headers", headers);
            initMessage.set("payload", payload);
            
            System.out.println("Sending connection init: " + initMessage.toString());
            client.send(initMessage.toString());
        } catch (Exception e) {
            System.err.println("Error sending connection init: " + e.getMessage());
        }
    }
    
    private void sendSubscription() {
        try {
            Thread.sleep(1000);
            
            ObjectNode subscriptionMessage = objectMapper.createObjectNode();
            subscriptionMessage.put("id", "1");
            subscriptionMessage.put("type", "start");
            
            ObjectNode payload = objectMapper.createObjectNode();
            payload.put("query", 
                "subscription UnitUpdates($unitId: String!) {\n" +
                "  unit(id: $unitId) {\n" +
                "    id\n" +
                "    measurements {\n" +
                "      ph\n" +
                "      rx\n" +
                "      cl\n" +
                "      temperature\n" +
                "      waterflow\n" +
                "    }\n" +
                "    variables {\n" +
                "      ph_setpoint\n" +
                "      rx_setpoint\n" +
                "      cl_setpoint\n" +
                "    }\n" +
                "    dosing {\n" +
                "      ph_minus\n" +
                "      cl\n" +
                "      floc\n" +
                "    }\n" +
                "  }\n" +
                "}"
            );
            
            ObjectNode variables = objectMapper.createObjectNode();
            variables.put("unitId", asekoConfig.getUnitId());
            payload.set("variables", variables);
            
            subscriptionMessage.set("payload", payload);
            
            System.out.println("Sending subscription: " + subscriptionMessage.toString());
            client.send(subscriptionMessage.toString());
        } catch (Exception e) {
            System.err.println("Error sending subscription: " + e.getMessage());
        }
    }
    
    private void handleDataMessage(JsonNode message) {
        try {
            JsonNode data = message.path("payload").path("data");
            if (!data.isMissingNode()) {
                JsonNode unit = data.path("unit");
                if (!unit.isMissingNode()) {
                    JsonNode measurements = unit.path("measurements");
                    if (!measurements.isMissingNode()) {
                        JsonNode waterflow = measurements.path("waterflow");
                        if (!waterflow.isMissingNode()) {
                            double waterflowValue = waterflow.asDouble();
                            boolean isFlowing = waterflowValue > 0;
                            
                            System.out.println("Waterflow value: " + waterflowValue + ", Flow status: " + isFlowing);
                            
                            // Update flow status and notify consumers
                            this.flowStatus.set(isFlowing);
                            if (flowStatusConsumer != null) {
                                flowStatusConsumer.accept(isFlowing);
                            }
                        }
                    }
                }
            }
        } catch (Exception e) {
            System.err.println("Error processing data message: " + e.getMessage());
        }
    }
} 