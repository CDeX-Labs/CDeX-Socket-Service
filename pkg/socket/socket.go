package socket

import (
    "net/http"
    "sync"
    "github.com/gorilla/websocket"
    "github.com/rs/zerolog/log"
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
        log.Error().Err(err).Msg("Error upgrading connection to websocket")
        return
    }
    defer conn.Close()

    log.Debug().Msg("Client has initiated a connection")

    clientsLock.Lock()
    clients[conn] = true
    clientsLock.Unlock()

    defer func() {
        clientsLock.Lock()
        delete(clients, conn)
        clientsLock.Unlock()
    }()

    go func() {
        for {
            _, msg, err := conn.ReadMessage()
            if err != nil {
                log.Error().Err(err).Msg("Error reading message")
                break
            }

            log.Debug().Str("message", string(msg)).Msg("Message received")

            clientsLock.RLock()
            for client := range clients {
                if client != conn {
                    err = client.WriteMessage(websocket.TextMessage, msg)
                    if err != nil {
                        log.Error().Err(err).Msg("Error writing message")
                        break
                    }
                }
            }
            clientsLock.RUnlock()
        }
    }()

    select {}
}
