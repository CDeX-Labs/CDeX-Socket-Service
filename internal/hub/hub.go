package hub

import (
	"encoding/json"
	"sync"

	"github.com/CDeX-Labs/CDeX-Socket-Service/pkg/protocol"
	"github.com/rs/zerolog"
)

type Hub struct {
	clients     map[*Client]bool
	userClients map[string]map[*Client]bool
	Register    chan *Client
	Unregister  chan *Client
	mu          sync.RWMutex
	logger      zerolog.Logger
	rooms       *RoomManager
}

func NewHub(logger zerolog.Logger) *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		userClients: make(map[string]map[*Client]bool),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		rooms:       NewRoomManager(),
		logger:      logger.With().Str("component", "hub").Logger(),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = true

	if h.userClients[client.UserID] == nil {
		h.userClients[client.UserID] = make(map[*Client]bool)
	}
	h.userClients[client.UserID][client] = true

	h.logger.Info().
		Str("clientId", client.ID).
		Str("userId", client.UserID).
		Int("totalClients", len(h.clients)).
		Msg("Client registered")
}

func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		h.rooms.LeaveAllRooms(client)

		delete(h.clients, client)
		close(client.Send)

		if userClients, ok := h.userClients[client.UserID]; ok {
			delete(userClients, client)
			if len(userClients) == 0 {
				delete(h.userClients, client.UserID)
			}
		}

		h.logger.Info().
			Str("clientId", client.ID).
			Str("userId", client.UserID).
			Int("totalClients", len(h.clients)).
			Msg("Client unregistered")
	}
}

func (h *Hub) ProcessMessage(client *Client, data []byte) {
	msg, err := protocol.ParseMessage(data)
	if err != nil {
		h.logger.Error().Err(err).Str("clientId", client.ID).Msg("Failed to parse message")
		h.sendError(client, "PARSE_ERROR", "Invalid message format", "")
		return
	}

	h.logger.Debug().
		Str("clientId", client.ID).
		Str("type", string(msg.Type)).
		Msg("Processing message")

	switch msg.Type {
	case protocol.MsgJoinRoom:
		h.handleJoinRoom(client, msg)
	case protocol.MsgLeaveRoom:
		h.handleLeaveRoom(client, msg)
	case protocol.MsgPing:
		h.handlePing(client, msg)
	default:
		h.sendError(client, "UNKNOWN_TYPE", "Unknown message type", msg.RequestID)
	}
}

func (h *Hub) handleJoinRoom(client *Client, msg *protocol.Message) {
	var payload protocol.JoinRoomPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		h.sendError(client, "INVALID_PAYLOAD", "Invalid join room payload", msg.RequestID)
		return
	}

	if payload.RoomID == "" {
		h.sendError(client, "INVALID_ROOM", "Room ID is required", msg.RequestID)
		return
	}

	room := h.rooms.JoinRoom(payload.RoomID, client)

	h.logger.Info().
		Str("clientId", client.ID).
		Str("roomId", payload.RoomID).
		Int("memberCount", room.ClientCount()).
		Msg("Client joined room")

	response, _ := protocol.NewMessageWithRequestID(protocol.MsgRoomJoined, protocol.RoomJoinedPayload{
		RoomID:      payload.RoomID,
		MemberCount: room.ClientCount(),
	}, msg.RequestID)

	h.SendToClient(client, response)
}

func (h *Hub) handleLeaveRoom(client *Client, msg *protocol.Message) {
	var payload protocol.LeaveRoomPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		h.sendError(client, "INVALID_PAYLOAD", "Invalid leave room payload", msg.RequestID)
		return
	}

	h.rooms.LeaveRoom(payload.RoomID, client)

	h.logger.Info().
		Str("clientId", client.ID).
		Str("roomId", payload.RoomID).
		Msg("Client left room")

	response, _ := protocol.NewMessageWithRequestID(protocol.MsgRoomLeft, protocol.RoomLeftPayload{
		RoomID: payload.RoomID,
	}, msg.RequestID)

	h.SendToClient(client, response)
}

func (h *Hub) handlePing(client *Client, msg *protocol.Message) {
	response, _ := protocol.NewMessageWithRequestID(protocol.MsgPong, nil, msg.RequestID)
	h.SendToClient(client, response)
}

func (h *Hub) SendToClient(client *Client, msg *protocol.Message) {
	data, err := msg.ToBytes()
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to serialize message")
		return
	}

	select {
	case client.Send <- data:
	default:
		h.logger.Warn().Str("clientId", client.ID).Msg("Client send buffer full, disconnecting")
		h.Unregister <- client
	}
}

func (h *Hub) SendToUser(userID string, msg *protocol.Message) {
	h.mu.RLock()
	clients := h.userClients[userID]
	h.mu.RUnlock()

	for client := range clients {
		h.SendToClient(client, msg)
	}
}

func (h *Hub) SendToRoom(roomID string, msg *protocol.Message) {
	room := h.rooms.GetRoom(roomID)
	if room == nil {
		return
	}

	data, err := msg.ToBytes()
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to serialize message")
		return
	}

	for _, client := range room.GetClients() {
		select {
		case client.Send <- data:
		default:
		}
	}
}

func (h *Hub) Broadcast(msg *protocol.Message) {
	data, err := msg.ToBytes()
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to serialize message")
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		select {
		case client.Send <- data:
		default:
		}
	}
}

func (h *Hub) sendError(client *Client, code, message, requestID string) {
	errMsg, _ := protocol.NewErrorMessage(code, message, requestID)
	h.SendToClient(client, errMsg)
}

func (h *Hub) getRoomMemberCount(roomID string) int {
	room := h.rooms.GetRoom(roomID)
	if room == nil {
		return 0
	}
	return room.ClientCount()
}

func (h *Hub) GetStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return map[string]interface{}{
		"totalClients": len(h.clients),
		"totalUsers":   len(h.userClients),
		"rooms":        h.rooms.GetStats(),
	}
}
