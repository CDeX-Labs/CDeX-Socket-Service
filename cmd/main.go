package main

import (
    //"fmt"
    "net/http"
    "Enigma-Socket-Service/pkg/socket"
    //"Enigma-Socket-Service/config"
    "github.com/rs/zerolog/log"
)

func main() {
    http.HandleFunc("/v1/ws", socket.HandleWebSocket)
    log.Info().Msg("Server started on port 6001")
    http.ListenAndServe(":6001", nil)
}

