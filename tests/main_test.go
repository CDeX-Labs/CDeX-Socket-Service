package main

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/rs/zerolog"
	"Enigma-Socket-Service/pkg/socket"
)

func TestMain(t *testing.T) {
	// Unit test for main function
    zerolog.SetGlobalLevel(zerolog.Disabled)

    req, err := http.NewRequest("GET", "/alerts", nil)
    if err != nil {
        t.Fatal(err)
    }

    rr := httptest.NewRecorder()

    socket.HandleWebSocket(rr, req)

    if status := rr.Code; status != http.StatusBadRequest {
        t.Errorf("handler returned wrong status code: got %v want %v",
            status, http.StatusBadRequest)
    }
}
