package socket

import (
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"net/http"
	"sync"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool)
var clientsLock = sync.RWMutex{}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade failed: ", err)
		return
	}
	defer conn.Close()

    // Echo incoming messages to all clients
    for {
        messageType, p, err := conn.ReadMessage()
        if err != nil {
            log.Debug().Msg("Error reading message")
            return
        }
        clientsLock.RLock()
        for client := range clients {
            if client != conn {
                if err := client.WriteMessage(messageType, p); err != nil {
                    log.Debug().Msg("Error writing message")
                    return
                }
            }
        }
        clientsLock.RUnlock()
    }
}
