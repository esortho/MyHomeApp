package com.example.asekoflowmonitor.service;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpGet;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.util.EntityUtils;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Profile;
import org.springframework.stereotype.Service;

import java.io.IOException;
import java.nio.charset.StandardCharsets;

@Service
@Profile("nodejs")
public class JavaScriptBridgeService {

    private final ObjectMapper objectMapper;
    
    @Value("${nodejs.server.url:http://localhost:3000}")
    private String nodeJsServerUrl;
    
    public JavaScriptBridgeService() {
        this.objectMapper = new ObjectMapper();
    }
    
    public boolean getFlowStatus() throws IOException {
        CloseableHttpClient httpClient = HttpClients.createDefault();
        HttpGet request = new HttpGet(nodeJsServerUrl + "/api/flowstatus");
        
        try (CloseableHttpResponse response = httpClient.execute(request)) {
            String responseBody = EntityUtils.toString(response.getEntity(), StandardCharsets.UTF_8);
            
            if (response.getStatusLine().getStatusCode() == 200) {
                JsonNode jsonResponse = objectMapper.readTree(responseBody);
                return jsonResponse.path("flowing").asBoolean(false);
            } else {
                System.err.println("Error getting flow status from Node.js server: " + response.getStatusLine());
                System.err.println("Response: " + responseBody);
                throw new IOException("Failed to get flow status from Node.js server");
            }
        }
    }
} 