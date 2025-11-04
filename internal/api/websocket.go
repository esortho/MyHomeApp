package api

import (
	"encoding/json"
	"fmt"
	"log"
	"myhomeapp/internal/config"

	"github.com/gorilla/websocket"
)

// WebSocketService handles real-time pool status updates
type WebSocketService struct {
	config     *config.Config
	client     *AsekoClient
	conn       *websocket.Conn
	done       chan struct{}
	statusChan chan bool
}

// NewWebSocketService creates a new WebSocket service
func NewWebSocketService(config *config.Config, client *AsekoClient) *WebSocketService {
	return &WebSocketService{
		config:     config,
		client:     client,
		done:       make(chan struct{}),
		statusChan: make(chan bool, 1),
	}
}

// Connect establishes a WebSocket connection
func (ws *WebSocketService) Connect() error {
	// Get the WebSocket URL from config
	wsURL := ws.config.Aseko.WebSocketURL
	if wsURL == "" {
		return fmt.Errorf("WebSocket URL not configured")
	}

	// Create WebSocket connection
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("error connecting to WebSocket: %w", err)
	}
	ws.conn = conn

	// Start reading messages
	go ws.readMessages()

	return nil
}

// readMessages reads messages from the WebSocket connection
func (ws *WebSocketService) readMessages() {
	defer close(ws.done)
	defer ws.conn.Close()

	for {
		_, message, err := ws.conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading WebSocket message: %v", err)
			return
		}

		var data struct {
			Type    string `json:"type"`
			Payload struct {
				FlowState bool `json:"flowState"`
			} `json:"payload"`
		}

		if err := json.Unmarshal(message, &data); err != nil {
			log.Printf("Error parsing WebSocket message: %v", err)
			continue
		}

		if data.Type == "flowStatus" {
			ws.statusChan <- data.Payload.FlowState
		}
	}
}

// GetStatusChan returns the channel for receiving flow status updates
func (ws *WebSocketService) GetStatusChan() <-chan bool {
	return ws.statusChan
}

// Close closes the WebSocket connection
func (ws *WebSocketService) Close() {
	if ws.conn != nil {
		ws.conn.Close()
	}
	close(ws.done)
}
