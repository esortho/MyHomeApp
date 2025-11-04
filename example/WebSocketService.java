package com.example.asekoflowmonitor.service;

import com.example.asekoflowmonitor.config.AsekoConfig;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.messaging.simp.stomp.StompCommand;
import org.springframework.messaging.simp.stomp.StompHeaders;
import org.springframework.messaging.simp.stomp.StompSession;
import org.springframework.messaging.simp.stomp.StompSessionHandler;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.web.socket.client.standard.StandardWebSocketClient;
import org.springframework.web.socket.messaging.WebSocketStompClient;
import org.springframework.web.socket.sockjs.client.SockJsClient;
import org.springframework.web.socket.sockjs.client.Transport;
import org.springframework.web.socket.sockjs.client.WebSocketTransport;

import javax.annotation.PostConstruct;
import java.lang.reflect.Type;
import java.util.ArrayList;
import java.util.List;
import java.util.function.Consumer;

@Service
public class WebSocketService {
    
    private final AsekoConfig asekoConfig;
    private final AuthService authService;
    private WebSocketStompClient stompClient;
    private StompSession stompSession;
    private Consumer<Boolean> flowStatusConsumer;
    private boolean wsEnabled = false; // Disable WebSocket initially until we understand the API
    
    @Autowired
    public WebSocketService(AsekoConfig asekoConfig, AuthService authService) {
        this.asekoConfig = asekoConfig;
        this.authService = authService;
    }
    
    @PostConstruct
    public void init() {
        if (!wsEnabled) {
            System.out.println("WebSocket is disabled for now");
            return;
        }
        
        try {
            List<Transport> transports = new ArrayList<>();
            transports.add(new WebSocketTransport(new StandardWebSocketClient()));
            SockJsClient sockJsClient = new SockJsClient(transports);
            
            this.stompClient = new WebSocketStompClient(sockJsClient);
        } catch (Exception e) {
            System.err.println("Error initializing WebSocket client: " + e.getMessage());
            e.printStackTrace();
        }
    }
    
    public void connect(Consumer<Boolean> flowStatusConsumer) {
        if (!wsEnabled || stompClient == null) {
            // Simulate periodic status updates since WebSocket is disabled
            this.flowStatusConsumer = flowStatusConsumer;
            return;
        }
        
        this.flowStatusConsumer = flowStatusConsumer;
        
        try {
            String token = authService.getAuthToken();
            System.out.println("Connecting to WebSocket at: " + asekoConfig.getWebsocketUrl());
            
            stompClient.connect(asekoConfig.getWebsocketUrl(), new StompSessionHandler() {
                @Override
                public void afterConnected(StompSession session, StompHeaders connectedHeaders) {
                    stompSession = session;
                    System.out.println("Connected to WebSocket. Session: " + session.getSessionId());
                    
                    // Subscribe to flow status updates
                    session.subscribe("/topic/flowStatus", this);
                    System.out.println("Subscribed to /topic/flowStatus");
                }
                
                @Override
                public void handleException(StompSession session, StompCommand command, StompHeaders headers, byte[] payload, Throwable exception) {
                    System.err.println("Error handling WebSocket command: " + command);
                    exception.printStackTrace();
                }
                
                @Override
                public void handleTransportError(StompSession session, Throwable exception) {
                    System.err.println("WebSocket transport error");
                    exception.printStackTrace();
                }
                
                @Override
                public Type getPayloadType(StompHeaders headers) {
                    return Boolean.class;
                }
                
                @Override
                public void handleFrame(StompHeaders headers, Object payload) {
                    System.out.println("Received WebSocket frame: " + payload);
                    if (payload instanceof Boolean) {
                        Boolean flowStatus = (Boolean) payload;
                        if (flowStatusConsumer != null) {
                            flowStatusConsumer.accept(flowStatus);
                        }
                    }
                }
            });
        } catch (Exception e) {
            System.err.println("Error connecting to WebSocket: " + e.getMessage());
            e.printStackTrace();
        }
    }
    
    @Scheduled(fixedRate = 30000) // Check every 30 seconds
    public void simulateStatusUpdates() {
        if (flowStatusConsumer != null) {
            // Generate random status for testing
            boolean status = Math.random() > 0.3; // 70% chance of true
            System.out.println("Simulating flow status update: " + status);
            flowStatusConsumer.accept(status);
        }
    }
    
    @Scheduled(fixedRate = 60000) // Reconnect every minute if disconnected
    public void checkConnection() {
        if (!wsEnabled) return;
        
        if (stompClient != null && (stompSession == null || !stompSession.isConnected())) {
            System.out.println("WebSocket disconnected. Attempting to reconnect...");
            connect(flowStatusConsumer);
        }
    }
} 