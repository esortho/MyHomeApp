package com.example.asekoflowmonitor.service;

import com.example.asekoflowmonitor.config.AsekoConfig;
import com.example.asekoflowmonitor.config.CredentialsConfig;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import org.java_websocket.client.WebSocketClient;
import org.java_websocket.handshake.ServerHandshake;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.stereotype.Service;

import javax.annotation.PostConstruct;
import java.net.URI;
import java.util.HashMap;
import java.util.Iterator;
import java.util.Map;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.function.Consumer;

@Service
public class DirectWebSocketService {

    private final AsekoConfig asekoConfig;
    private final AuthService authService;
    private final ObjectMapper objectMapper;
    private final SimpMessagingTemplate messagingTemplate;
    private WebSocketClient client;
    private AtomicBoolean flowStatus = new AtomicBoolean(false);
    private Consumer<Boolean> flowStatusConsumer;

    @Autowired
    public DirectWebSocketService(AsekoConfig asekoConfig, 
                                 AuthService authService,
                                 SimpMessagingTemplate messagingTemplate) {
        this.asekoConfig = asekoConfig;
        this.authService = authService;
        this.objectMapper = new ObjectMapper();
        this.messagingTemplate = messagingTemplate;
    }

    @PostConstruct
    public void init() {
        try {
            // Try authentication first
            authService.login();
            
            // Then give time for authentication to complete
            Thread.sleep(2000);
            
            // Check if authentication was successful
            if (authService.isAuthenticated()) {
                System.out.println("Authentication successful, connecting to WebSocket...");
                connectWebSocket(status -> {
                    this.flowStatus.set(status);
                    System.out.println("Flow status updated: " + status);
                });
            } else {
                System.err.println("Cannot connect to WebSocket: Not authenticated");
            }
        } catch (Exception e) {
            System.err.println("Failed to initialize WebSocket: " + e.getMessage());
            e.printStackTrace();
        }
    }

    public void connectWebSocket(Consumer<Boolean> consumer) {
        try {
            String token = authService.getAuthToken();
            this.flowStatusConsumer = consumer;
            
            System.out.println("Connecting directly to GraphQL WebSocket: " + asekoConfig.getGraphqlWsUrl());
            
            // Set up headers
            Map<String, String> headers = new HashMap<>();
            headers.put("Authorization", "Bearer " + token);
            headers.put("Accept", "*/*");
            headers.put("Accept-Language", "en-GB,en;q=0.9");
            headers.put("Origin", "https://aseko.cloud");
            headers.put("Referer", "https://aseko.cloud/");
            headers.put("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36");
            
            this.client = new WebSocketClient(new URI(asekoConfig.getGraphqlWsUrl()), headers) {
                @Override
                public void onOpen(ServerHandshake handshake) {
                    System.out.println("WebSocket connection established");
                    System.out.println("HTTP Status: " + handshake.getHttpStatus());
                    System.out.println("HTTP Status Message: " + handshake.getHttpStatusMessage());
                    
                    // Print all headers
                    System.out.println("Headers:");
                    for (Iterator<String> it = handshake.iterateHttpFields(); it.hasNext();) {
                        String key = it.next();
                        System.out.println("  " + key + ": " + handshake.getFieldValue(key));
                    }
                    
                    // Now send the connection init message
                    sendConnectionInit(token);
                }

                @Override
                public void onMessage(String message) {
                    try {
                        System.out.println("WebSocket message received: " + message);
                        JsonNode messageJson = objectMapper.readTree(message);
                        String type = messageJson.path("type").asText();
                        
                        if ("connection_ack".equals(type)) {
                            System.out.println("Connection acknowledged, sending subscription");
                            sendSubscription();
                        } else if ("data".equals(type)) {
                            handleDataMessage(messageJson);
                        }
                    } catch (Exception e) {
                        System.err.println("Error processing WebSocket message: " + e.getMessage());
                    }
                }

                @Override
                public void onClose(int code, String reason, boolean remote) {
                    System.out.println("WebSocket connection closed: " + reason);
                }

                @Override
                public void onError(Exception ex) {
                    System.err.println("WebSocket error: " + ex.getMessage());
                }
            };
            
            this.client.connect();
            
        } catch (Exception e) {
            System.err.println("Error connecting to WebSocket: " + e.getMessage());
            e.printStackTrace();
        }
    }
    
    private void sendConnectionInit(String token) {
        try {
            System.out.println("Sending connection init message");
            
            ObjectNode initMessage = objectMapper.createObjectNode();
            initMessage.put("type", "connection_init");
            
            ObjectNode payload = objectMapper.createObjectNode();
            ObjectNode headers = objectMapper.createObjectNode();
            headers.put("Authorization", "Bearer " + token);
            headers.put("X-App-Name", "pool-live");
            headers.put("X-App-Version", "4.2.0");
            headers.put("X-Mode", "production");
            headers.put("X-Cloud", asekoConfig.getCloudId());
            
            payload.set("headers", headers);
            initMessage.set("payload", payload);
            
            String initMessageString = objectMapper.writeValueAsString(initMessage);
            System.out.println("Connection init message: " + initMessageString);
            client.send(initMessageString);
            
        } catch (Exception e) {
            System.err.println("Error sending connection init: " + e.getMessage());
        }
    }
    
    private void sendSubscription() {
        try {
            System.out.println("Sending subscription message");
            
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
            
            client.send(objectMapper.writeValueAsString(subscriptionMessage));
            
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
                    // Forward the complete unit data to clients
                    messagingTemplate.convertAndSend("/topic/poolData", unit);
                    
                    JsonNode measurements = unit.path("measurements");
                    if (!measurements.isMissingNode()) {
                        JsonNode waterflow = measurements.path("waterflow");
                        if (!waterflow.isMissingNode()) {
                            double waterflowValue = waterflow.asDouble();
                            boolean isFlowing = waterflowValue > 0;
                            
                            System.out.println("Waterflow value: " + waterflowValue + ", Flow status: " + isFlowing);
                            
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
    
    public boolean getFlowStatus() {
        return flowStatus.get();
    }
} 