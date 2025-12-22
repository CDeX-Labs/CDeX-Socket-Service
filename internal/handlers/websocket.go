package handlers

import (
	"net/http"

	"github.com/CDeX-Labs/CDeX-Socket-Service/internal/auth"
	"github.com/CDeX-Labs/CDeX-Socket-Service/internal/hub"
	"github.com/CDeX-Labs/CDeX-Socket-Service/internal/presence"
	"github.com/CDeX-Labs/CDeX-Socket-Service/pkg/protocol"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketHandler struct {
	hub       *hub.Hub
	presence  *presence.Manager
	logger    zerolog.Logger
}

func NewWebSocketHandler(h *hub.Hub, p *presence.Manager, logger zerolog.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		hub:       h,
		presence:  p,
		logger:    logger.With().Str("component", "ws-handler").Logger(),
	}
}

func (h *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to upgrade connection")
		return
	}

	clientID := uuid.New().String()
	userID := claims.GetUserID()

	client := hub.NewClient(clientID, userID, conn, h.hub, h.logger)

	h.hub.Register <- client

	if h.presence != nil {
		h.presence.SetOnline(r.Context(), userID)
	}

	connectedMsg, _ := protocol.NewMessage(protocol.MsgConnected, protocol.ConnectedPayload{
		UserID:     userID,
		InstanceID: clientID,
	})
	h.hub.SendToClient(client, connectedMsg)

	h.logger.Info().
		Str("clientId", clientID).
		Str("userId", userID).
		Str("remoteAddr", r.RemoteAddr).
		Msg("WebSocket connection established")

	go client.WritePump()
	go client.ReadPump()
}

func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func ReadyHandler(h *hub.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		stats := h.GetStats()
		w.Write([]byte(`{"status":"ready","stats":` + toJSON(stats) + `}`))
	}
}

func toJSON(v interface{}) string {
	switch val := v.(type) {
	case map[string]interface{}:
		result := "{"
		first := true
		for k, v := range val {
			if !first {
				result += ","
			}
			result += `"` + k + `":` + toJSON(v)
			first = false
		}
		result += "}"
		return result
	case int:
		return string(rune(val + '0'))
	default:
		return `"unknown"`
	}
}
