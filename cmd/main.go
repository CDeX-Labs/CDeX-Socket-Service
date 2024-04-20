package main

import (
    "net/http"
    "Enigma-Socket-Service/pkg/socket"
    "github.com/rs/zerolog/log"
)

func main() {
    http.HandleFunc("/alerts", socket.HandleWebSocket)
    log.Info().Msg("Server started on port 6001")
    http.ListenAndServe(":6001", nil)
}

