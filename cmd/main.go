package main

import (
    "fmt"
    "net/http"
    "Enigma-Socket-Service/pkg/socket"
    "Enigma-Socket-Service/config"
    "github.com/rs/zerolog/log"
)

func main() {
    http.HandleFunc("/alerts", socket.HandleWebSocket)
    log.Info().Msg("Server started on port 6001")
    http.ListenAndServe(fmt.Sprintf(":%s", config.Config.App.Port), nil)
}

