package socket

import (
    "net/http"
    "github.com/gorilla/websocket"
    "github.com/rs/zerolog/log"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Error().Err(err).Msg("Error upgrading connection to websocket")
        return
    }
    defer conn.Close()

    log.Debug().Msg("Client has initiated a connection")

    go func() {
        for {
            _, msg, err := conn.ReadMessage()
            if err != nil {
                log.Error().Err(err).Msg("Error reading message")
                break
            }
            
            log.Debug().Str("message", string(msg)).Msg("Message received")

            err = conn.WriteMessage(websocket.TextMessage, msg)
            if err != nil {
                log.Error().Err(err).Msg("Error writing message")
                break
            }
        }
    }()

    select {}
}