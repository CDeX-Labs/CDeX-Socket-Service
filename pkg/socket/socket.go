package socket

import (
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"net/http"
	"sync"
	"time"
)

var apiKeys = []string{
	"choco",
	"arsh"
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
	for _, validKey := range apiKeys {
		if key == validKey {
			return true
		}
	}
	return false
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	apiKey := r.URL.Query().Get("apiKey")
	if !isValidAPIKey(apiKey) {
		http.Error(w, "Invalid API Key", http.StatusUnauthorized)
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

		// Broadcast message to all clients
		connections.RLock()
		for client := range connections.clients {
			err := client.WriteMessage(messageType, message)
			if err != nil {
				log.Error().Err(err).Msg("Error during message writing")
				client.Close()
				delete(connections.clients, client)
			}
		}
		connections.RUnlock()
	}
}
