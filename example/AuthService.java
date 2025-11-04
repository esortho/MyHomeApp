package com.example.asekoflowmonitor.service;

import com.example.asekoflowmonitor.config.AsekoConfig;
import com.example.asekoflowmonitor.config.CredentialsConfig;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.apache.http.client.CookieStore;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpGet;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.cookie.Cookie;
import org.apache.http.entity.ContentType;
import org.apache.http.entity.StringEntity;
import org.apache.http.impl.client.BasicCookieStore;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.util.EntityUtils;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.stereotype.Service;

import javax.annotation.PostConstruct;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.util.List;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

@Service
public class AuthService {
    
    private final AsekoConfig asekoConfig;
    private final CredentialsConfig credentialsConfig;
    private final ObjectMapper objectMapper;
    private final CookieStore cookieStore = new BasicCookieStore();
    private final SimpMessagingTemplate messagingTemplate;
    
    private String authToken;
    private String refreshToken;
    private String userId;
    private boolean isAuthenticated = false;
    
    private JsonNode userProfile;
    
    // Aseko Cloud specific constants
    private static final String AUTH_ENDPOINT = "https://auth.aseko.acs.aseko.cloud/auth/login";
    private static final String CLOUD_ID = "01HXS50KTV7NRSVNHD617J4CKB"; // This might be user-specific
    
    @Autowired
    public AuthService(AsekoConfig asekoConfig, CredentialsConfig credentialsConfig, SimpMessagingTemplate messagingTemplate) {
        this.asekoConfig = asekoConfig;
        this.credentialsConfig = credentialsConfig;
        this.objectMapper = new ObjectMapper();
        this.messagingTemplate = messagingTemplate;
    }
    
    // This method reads the password more safely
    private String getCredentialsPassword() {
        // First try to get from config
        String configPassword = credentialsConfig.getPassword();
        
        // Log what we found (for debugging)
        System.out.println("Password from config: " + (configPassword != null ? 
                          (configPassword.isEmpty() ? "empty" : "******") : "null"));
        
        // If not available, use the hard-coded fallback
        if (configPassword == null || configPassword.isEmpty()) {
            return "!Saab939";  // Fallback password
        }
        
        return configPassword;
    }
    
    @PostConstruct
    public void init() {
        // Just log configuration at startup
        System.out.println("Credentials loaded - Email: " + credentialsConfig.getEmail());
        System.out.println("Password available: " + (getCredentialsPassword() != null));
    }
    
    public boolean isAuthenticated() {
        return isAuthenticated;
    }
    
    public String getAuthToken() {
        if (authToken == null) {
            try {
                login();
            } catch (Exception e) {
                System.err.println("Failed to get auth token: " + e.getMessage());
            }
        }
        return authToken;
    }
    
    public List<Cookie> getCookies() {
        return cookieStore.getCookies();
    }
    
    public CloseableHttpClient getHttpClient() {
        return HttpClients.custom()
                .setDefaultCookieStore(cookieStore)
                .build();
    }
    
    public boolean login() throws IOException {
        System.out.println("\n===== DIRECT LOGIN ATTEMPT =====");
        
        try {
            // Create the HttpClient exactly like in the working test
            CloseableHttpClient httpClient = HttpClients.createDefault();
            
            // Create HttpPost with the exact URL that works
            HttpPost httpPost = new HttpPost("https://auth.aseko.acs.aseko.cloud/auth/login");
            
            // Set all headers exactly as in the working test
            httpPost.setHeader("Accept", "application/json");
            httpPost.setHeader("Accept-Language", "en");
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
            httpPost.setHeader("sec-ch-ua", "\"Not(A:Brand\";v=\"99\", \"Google Chrome\";v=\"133\", \"Chromium\";v=\"133\"");
            httpPost.setHeader("sec-ch-ua-mobile", "?0");
            httpPost.setHeader("sec-ch-ua-platform", "\"macOS\"");
            
            // Create the JSON payload using a hardcoded string first
            // This is to match EXACTLY what works in the test
            String jsonPayload = "{\"email\":\"soren.thornlund@gmail.com\",\"password\":\"!Saab939\",\"cloud\":\"01HXS50KTV7NRSVNHD617J4CKB\"}";
            
            // Mask password in logs
            String maskedPayload = jsonPayload.replaceAll("\"password\":\"[^\"]*\"", "\"password\":\"********\"");
            System.out.println("Direct login payload: " + maskedPayload);
            
            // Create StringEntity with ContentType.APPLICATION_JSON - exactly like the working test
            StringEntity entity = new StringEntity(jsonPayload, ContentType.APPLICATION_JSON);
            httpPost.setEntity(entity);
            
            // Execute the request
            CloseableHttpResponse response = httpClient.execute(httpPost);
            try {
                int statusCode = response.getStatusLine().getStatusCode();
                String responseBody = EntityUtils.toString(response.getEntity());
                
                System.out.println("Direct login response: " + statusCode);
                
                if (statusCode == 200) {
                    // Parse the token from the successful response
                    JsonNode responseJson = objectMapper.readTree(responseBody);
                    this.authToken = responseJson.path("token").asText();
                    this.isAuthenticated = true;
                    
                    System.out.println("Login successful! Token: " + 
                        (this.authToken.length() > 10 ? this.authToken.substring(0, 10) + "..." : this.authToken));
                } else {
                    System.err.println("Login failed. Status: " + statusCode);
                    System.err.println("Response: " + responseBody);
                    throw new IOException("Login failed with status: " + statusCode);
                }
            } finally {
                response.close();
                httpClient.close();
            }
        } catch (Exception e) {
            System.err.println("Login error: " + e.getMessage());
            e.printStackTrace();
            throw new IOException("Login failed: " + e.getMessage(), e);
        }
        
        System.out.println("===== DIRECT LOGIN ATTEMPT COMPLETE =====\n");
        
        // After successful login, fetch user profile
        if (isAuthenticated()) {
            try {
                fetchUserProfile();
            } catch (Exception e) {
                System.err.println("Failed to fetch user profile: " + e.getMessage());
            }
        }
        
        return isAuthenticated();
    }
    
    private void fetchUserInfo() {
        try {
            System.out.println("Fetching user information...");
            
            CloseableHttpClient httpClient = getHttpClient();
            HttpGet httpGet = new HttpGet("https://auth.aseko.acs.aseko.cloud/auth/me");
            
            // Set headers for user info request
            httpGet.setHeader("Accept", "application/json");
            httpGet.setHeader("Accept-Language", "en");
            httpGet.setHeader("Origin", "https://aseko.cloud");
            httpGet.setHeader("Referer", "https://aseko.cloud/");
            httpGet.setHeader("Authorization", "Bearer " + authToken);
            httpGet.setHeader("X-App-Name", "pool-live");
            httpGet.setHeader("X-App-Version", "4.2.0");
            httpGet.setHeader("X-Mode", "production");
            
            try (CloseableHttpResponse response = httpClient.execute(httpGet)) {
                int statusCode = response.getStatusLine().getStatusCode();
                String responseBody = EntityUtils.toString(response.getEntity(), StandardCharsets.UTF_8);
                
                if (statusCode == 200) {
                    JsonNode userInfo = objectMapper.readTree(responseBody);
                    JsonNode idNode = userInfo.path("id");
                    JsonNode nameNode = userInfo.path("name");
                    
                    if (!idNode.isMissingNode()) {
                        this.userId = idNode.asText();
                    }
                    
                    System.out.println("User info retrieved successfully");
                    if (!nameNode.isMissingNode()) {
                        System.out.println("Logged in as: " + nameNode.asText());
                    }
                } else {
                    System.out.println("Warning: Could not fetch user info, status: " + statusCode);
                }
            }
        } catch (Exception e) {
            System.out.println("Warning: Error fetching user info: " + e.getMessage());
            // Don't throw here, as we already have the auth token
        }
    }
    
    private String extractFormAction(String html) {
        Pattern pattern = Pattern.compile("<form[^>]*action=['\"]([^'\"]*)['\"]", Pattern.CASE_INSENSITIVE);
        Matcher matcher = pattern.matcher(html);
        if (matcher.find()) {
            return matcher.group(1);
        }
        return null;
    }
    
    private String extractCsrfToken(String html) {
        Pattern pattern = Pattern.compile("<input[^>]*name=['\"](_csrf|csrf_token|_token)['\"][^>]*value=['\"](.*?)['\"]", 
                                        Pattern.CASE_INSENSITIVE);
        Matcher matcher = pattern.matcher(html);
        if (matcher.find()) {
            return matcher.group(2);
        }
        return null;
    }
    
    private String extractApiEndpoint(String html) {
        // Look for API endpoints in JavaScript
        Pattern pattern = Pattern.compile("(api\\/login|auth\\/login|login\\.do|loginAction|authenticate)['\"]", 
                                        Pattern.CASE_INSENSITIVE);
        Matcher matcher = pattern.matcher(html);
        if (matcher.find()) {
            String endpoint = matcher.group(1);
            return asekoConfig.getBaseUrl() + "/" + endpoint;
        }
        return null;
    }
    
    // Store user profile data
    public void setUserProfile(JsonNode profile) {
        this.userProfile = profile;
        System.out.println("User profile updated: " + profile.path("name").asText());
        
        // Broadcast user profile to clients
        messagingTemplate.convertAndSend("/topic/userProfile", profile);
    }
    
    // Get user profile data
    public JsonNode getUserProfile() {
        return userProfile;
    }
    
    // Add method to fetch user profile
    private void fetchUserProfile() throws IOException {
        System.out.println("Fetching user profile...");
        
        try (CloseableHttpClient httpClient = HttpClients.createDefault()) {
            HttpGet httpGet = new HttpGet("https://api.acs.prod.aseko.cloud/users/me");
            
            // Set headers
            httpGet.setHeader("Accept", "application/json");
            httpGet.setHeader("Authorization", "Bearer " + getAuthToken());
            
            try (CloseableHttpResponse response = httpClient.execute(httpGet)) {
                int statusCode = response.getStatusLine().getStatusCode();
                String responseBody = EntityUtils.toString(response.getEntity(), StandardCharsets.UTF_8);
                
                if (statusCode == 200) {
                    JsonNode profile = objectMapper.readTree(responseBody);
                    setUserProfile(profile);
                    System.out.println("User profile fetched successfully");
                } else {
                    System.err.println("Failed to fetch profile. Status: " + statusCode);
                    System.err.println("Response: " + responseBody);
                }
            }
        } catch (Exception e) {
            System.err.println("Error fetching user profile: " + e.getMessage());
            throw new IOException("Failed to fetch user profile", e);
        }
    }
} 