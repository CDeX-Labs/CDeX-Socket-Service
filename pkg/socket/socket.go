package socket

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type APIKeyDetails struct {
	APIKey     string
	Username   string
	CreatedAt  string
}

var validAPIKeys = []APIKeyDetails{
	{
		APIKey:    "eofqjvwilceiecjwqludjtshrcnfcduo",
		Username:  "enigma",
		CreatedAt: "2024-08-07T17:20:31.087Z",
	},
	{
		APIKey:    "lursymyslygmjrmvuvzieekakkogjo",
		Username:  "martin",
		CreatedAt: "2024-08-07T17:20:42.236Z",
	},
}

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			if r.Host != "ws.enigma.fm" {
				return false
			} else {
				return true
			}
		},
	}
	connections = struct {
		sync.RWMutex
		clients map[*websocket.Conn]struct{}
	}{clients: make(map[*websocket.Conn]struct{})}
)

func isValidAPIKey(key string) bool {
	for _, apiKeyDetails := range validAPIKeys {
		if key == apiKeyDetails.APIKey {
			return true
		}
	}
	return false
}

func getLogFilePath(apiKey string) string {
	return filepath.Join("logs", fmt.Sprintf("%s.log", apiKey))
}

func createLogsDir() error {
	return os.MkdirAll("logs", os.ModePerm)
}

func logMessage(apiKey, messageType, message string) error {
	filePath := getLogFilePath(apiKey)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	logEntry := fmt.Sprintf("%s (%s): %s\n", messageType, time.Now().Format(time.RFC3339), message)
	_, err = file.WriteString(logEntry)
	return err
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	apiKey := r.URL.Query().Get("apiKey")
	if !isValidAPIKey(apiKey) {
		http.Error(w, "Invalid API Key", http.StatusUnauthorized)
		return
	}

	err := createLogsDir()
	if err != nil {
		log.Error().Err(err).Msg("Error creating logs directory")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("Error during connection upgradation")
		return
	}
	defer func() {
		connections.Lock()
		delete(connections.clients, conn)
		connections.Unlock()
	}()

	connections.Lock()
	connections.clients[conn] = struct{}{}
	connections.Unlock()

	// Periodic ping to keep connection alive
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Error().Err(err).Msg("Error sending ping")
					return
				}
			}
		}
	}()

	// Handle incoming messages
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Error().Err(err).Msg("Error during message reading")
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				return
			}
			break
		}
		log.Debug().Msgf("Received message: %s", message)

		// Log received message
		err = logMessage(apiKey, "RECEIVED", string(message))
		if err != nil {
			log.Error().Err(err).Msg("Error logging received message")
		}

		// Broadcast message to all clients
		connections.RLock()
		for client := range connections.clients {
			err := client.WriteMessage(messageType, message)
			if err != nil {
				log.Error().Err(err).Msg("Error during message writing")
				client.Close()
				delete(connections.clients, client)
			} else {
				// Log sent message
				err = logMessage(apiKey, "SENT", string(message))
				if err != nil {
					log.Error().Err(err).Msg("Error logging sent message")
				}
			}
		}
		connections.RUnlock()
	}
}