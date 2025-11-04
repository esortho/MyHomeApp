package com.example.asekoflowmonitor.controller;

import com.example.asekoflowmonitor.service.DirectWebSocketService;
import com.example.asekoflowmonitor.service.GraphQLService;
import com.example.asekoflowmonitor.service.UnitService;
import com.example.asekoflowmonitor.service.WebSocketService;
import com.fasterxml.jackson.databind.JsonNode;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.stereotype.Controller;
import org.springframework.ui.Model;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.ResponseBody;
import com.example.asekoflowmonitor.service.AuthService;

@Controller
public class FlowStatusController {

    private final GraphQLService graphQLService;
    private final WebSocketService webSocketService;
    private final SimpMessagingTemplate messagingTemplate;
    private final DirectWebSocketService directWebSocketService;
    private final UnitService unitService;
    private final AuthService authService;
    private boolean currentFlowStatus = false;
    
    @Autowired
    public FlowStatusController(GraphQLService graphQLService, 
                                WebSocketService webSocketService,
                                SimpMessagingTemplate messagingTemplate,
                                DirectWebSocketService directWebSocketService,
                                UnitService unitService,
                                AuthService authService) {
        this.graphQLService = graphQLService;
        this.webSocketService = webSocketService;
        this.messagingTemplate = messagingTemplate;
        this.directWebSocketService = directWebSocketService;
        this.unitService = unitService;
        this.authService = authService;
        
        // Register callback for WebSocket updates - using the direct service now
        this.directWebSocketService.connectWebSocket(this::updateFlowStatus);
    }
    
    @GetMapping("/")
    public String index(Model model) {
        try {
            // Initial status from DirectWebSocketService
            boolean status = directWebSocketService.getFlowStatus();
            model.addAttribute("flowStatus", status);
            
            // Add unitList data if available
            JsonNode unitListData = unitService.getUnitList();
            if (unitListData != null) {
                model.addAttribute("unitList", unitListData.toString());
            }
        } catch (Exception e) {
            e.printStackTrace();
            model.addAttribute("flowStatus", false);
            model.addAttribute("error", e.getMessage());
        }
        
        return "index";
    }
    
    @GetMapping("/api/status")
    @ResponseBody
    public boolean getStatus() {
        try {
            return directWebSocketService.getFlowStatus();
        } catch (Exception e) {
            e.printStackTrace();
            return false;
        }
    }
    
    @GetMapping("/api/units")
    @ResponseBody
    public JsonNode getUnits() {
        return unitService.getUnitList();
    }
    
    @GetMapping("/api/selected-unit")
    @ResponseBody
    public JsonNode getSelectedUnit() {
        return unitService.getSelectedUnit();
    }
    
    @GetMapping("/api/unit/{serialNumber}")
    @ResponseBody
    public ResponseEntity<JsonNode> getUnitDetails(@PathVariable String serialNumber) {
        try {
            JsonNode unitDetail = unitService.fetchUnitDetail(serialNumber);
            if (unitDetail != null) {
                return ResponseEntity.ok(unitDetail);
            } else {
                return ResponseEntity.notFound().build();
            }
        } catch (Exception e) {
            e.printStackTrace();
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }
    
    @GetMapping("/api/user-profile")
    @ResponseBody
    public JsonNode getUserProfile() {
        return authService.getUserProfile();
    }
    
    private void updateFlowStatus(boolean status) {
        this.currentFlowStatus = status;
        // Send update to connected clients
        messagingTemplate.convertAndSend("/topic/status", status);
        System.out.println("Sent flow status update to clients: " + status);
    }
} 