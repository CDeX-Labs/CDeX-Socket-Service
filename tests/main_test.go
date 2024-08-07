package main

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/rs/zerolog"
    "Enigma-Socket-Service/pkg/socket"
)

func TestMain(t *testing.T) {
    zerolog.SetGlobalLevel(zerolog.Disabled)

    // Test case with a valid API key
    validAPIKey := "eofqjvwilceiecjwqludjtshrcnfcduo"
    reqValid, err := http.NewRequest("GET", "/ws/v1?apiKey="+validAPIKey, nil)
    if err != nil {
        t.Fatal(err)
    }

    rrValid := httptest.NewRecorder()
    socket.HandleWebSocket(rrValid, reqValid)

    if status := rrValid.Code; status != http.StatusBadRequest {
        t.Errorf("handler returned wrong status code: got %v want %v",
            status, http.StatusBadRequest)
    }

    // Test case with an invalid API key
    invalidAPIKey := "INVALID_API_KEY"
    reqInvalid, err := http.NewRequest("GET", "/ws/v1?apiKey="+invalidAPIKey, nil)
    if err != nil {
        t.Fatal(err)
    }

    rrInvalid := httptest.NewRecorder()
    socket.HandleWebSocket(rrInvalid, reqInvalid)

    if status := rrInvalid.Code; status != http.StatusUnauthorized {
        t.Errorf("handler returned wrong status code: got %v want %v",
            status, http.StatusUnauthorized)
    }
}
