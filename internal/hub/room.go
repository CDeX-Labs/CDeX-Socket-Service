package hub

import (
	"strings"
	"sync"
	"time"
)

type RoomType string

const (
	RoomTypeGlobal  RoomType = "global"
	RoomTypeContest RoomType = "contest"
	RoomTypeProblem RoomType = "problem"
	RoomTypeUser    RoomType = "user"
)

type Room struct {
	ID        string
	Type      RoomType
	CreatedAt time.Time

	clients map[*Client]bool
	mu      sync.RWMutex
}

func NewRoom(id string) *Room {
	return &Room{
		ID:        id,
		Type:      ParseRoomType(id),
		CreatedAt: time.Now(),
		clients:   make(map[*Client]bool),
	}
}

func ParseRoomType(roomID string) RoomType {
	if roomID == "global" {
		return RoomTypeGlobal
	}

	parts := strings.SplitN(roomID, ":", 2)
	if len(parts) != 2 {
		return RoomTypeGlobal
	}

	switch parts[0] {
	case "contest":
		return RoomTypeContest
	case "problem":
		return RoomTypeProblem
	case "user":
		return RoomTypeUser
	default:
		return RoomTypeGlobal
	}
}

func ExtractRoomEntityID(roomID string) string {
	parts := strings.SplitN(roomID, ":", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return roomID
}

func BuildRoomID(roomType RoomType, entityID string) string {
	if roomType == RoomTypeGlobal {
		return "global"
	}
	return string(roomType) + ":" + entityID
}

func (r *Room) AddClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[client] = true
}

func (r *Room) RemoveClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clients, client)
}

func (r *Room) HasClient(client *Client) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.clients[client]
}

func (r *Room) GetClients() []*Client {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clients := make([]*Client, 0, len(r.clients))
	for client := range r.clients {
		clients = append(clients, client)
	}
	return clients
}

func (r *Room) ClientCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients)
}

func (r *Room) IsEmpty() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients) == 0
}

type RoomManager struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*Room),
	}
}

func (rm *RoomManager) GetOrCreateRoom(roomID string) *Room {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if room, exists := rm.rooms[roomID]; exists {
		return room
	}

	room := NewRoom(roomID)
	rm.rooms[roomID] = room
	return room
}

func (rm *RoomManager) GetRoom(roomID string) *Room {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.rooms[roomID]
}

func (rm *RoomManager) RemoveRoom(roomID string) bool {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return false
	}

	if room.IsEmpty() {
		delete(rm.rooms, roomID)
		return true
	}
	return false
}

func (rm *RoomManager) JoinRoom(roomID string, client *Client) *Room {
	room := rm.GetOrCreateRoom(roomID)
	room.AddClient(client)
	client.JoinRoom(roomID) // Also track in client
	return room
}

func (rm *RoomManager) LeaveRoom(roomID string, client *Client) {
	rm.mu.RLock()
	room := rm.rooms[roomID]
	rm.mu.RUnlock()

	if room != nil {
		room.RemoveClient(client)
		client.LeaveRoom(roomID)

		if room.IsEmpty() && room.Type != RoomTypeGlobal {
			rm.RemoveRoom(roomID)
		}
	}
}

func (rm *RoomManager) LeaveAllRooms(client *Client) {
	rooms := client.GetRooms()
	for _, roomID := range rooms {
		rm.LeaveRoom(roomID, client)
	}
}

func (rm *RoomManager) GetRoomsByType(roomType RoomType) []*Room {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	var result []*Room
	for _, room := range rm.rooms {
		if room.Type == roomType {
			result = append(result, room)
		}
	}
	return result
}

func (rm *RoomManager) GetStats() map[string]interface{} {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	typeCount := make(map[RoomType]int)
	totalClients := 0

	for _, room := range rm.rooms {
		typeCount[room.Type]++
		totalClients += room.ClientCount()
	}

	return map[string]interface{}{
		"totalRooms":   len(rm.rooms),
		"totalClients": totalClients,
		"byType":       typeCount,
	}
}
