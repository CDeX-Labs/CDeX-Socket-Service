package socket

import (
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"net/http"
	"sync"
)

var (
	upgrader = websocket.Upgrader{
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

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error during connection upgradation: ", err)
		return
	}
	defer func() {
		conn.Close()
		connections.Lock()
		delete(connections.clients, conn)
		connections.Unlock()
	}()

	connections.Lock()
	connections.clients[conn] = struct{}{}
	connections.Unlock()

	// Echo incoming messages to all clients
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Err(err).Msg("Error during message reading")
			break
		}
		log.Debug().Msgf("Received message: %s", message)

		// Broadcast message to all clients
		connections.RLock()
		for client := range connections.clients {
			err := client.WriteMessage(messageType, message)
			if err != nil {
				log.Err(err).Msg("Error during message writing")
				client.Close()
				delete(connections.clients, client)
			}
		}
		connections.RUnlock()
	}
}
