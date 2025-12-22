package protocol

import (
	"encoding/json"
	"time"
)

type MessageType string

const (
	MsgJoinRoom    MessageType = "JOIN_ROOM"
	MsgLeaveRoom   MessageType = "LEAVE_ROOM"
	MsgPing        MessageType = "PING"
	MsgSubscribe   MessageType = "SUBSCRIBE"
	MsgUnsubscribe MessageType = "UNSUBSCRIBE"

	MsgSubmissionCreated MessageType = "SUBMISSION_CREATED"
	MsgSubmissionResult  MessageType = "SUBMISSION_RESULT"
	MsgLeaderboardUpdate MessageType = "LEADERBOARD_UPDATE"
	MsgContestEvent      MessageType = "CONTEST_EVENT"
	MsgPresenceUpdate    MessageType = "PRESENCE_UPDATE"
	MsgRoomJoined        MessageType = "ROOM_JOINED"
	MsgRoomLeft          MessageType = "ROOM_LEFT"
	MsgPong              MessageType = "PONG"
	MsgError             MessageType = "ERROR"
	MsgConnected         MessageType = "CONNECTED"
)

type Message struct {
	Type      MessageType     `json:"type"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	Timestamp int64           `json:"timestamp"`
	RequestID string          `json:"requestId,omitempty"`
}

func NewMessage(msgType MessageType, payload interface{}) (*Message, error) {
	var payloadBytes json.RawMessage
	var err error

	if payload != nil {
		payloadBytes, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
	}

	return &Message{
		Type:      msgType,
		Payload:   payloadBytes,
		Timestamp: time.Now().UnixMilli(),
	}, nil
}

func NewMessageWithRequestID(msgType MessageType, payload interface{}, requestID string) (*Message,
	error) {
	msg, err := NewMessage(msgType, payload)
	if err != nil {
		return nil, err
	}
	msg.RequestID = requestID
	return msg, nil
}

func (m *Message) ToBytes() ([]byte, error) {
	return json.Marshal(m)
}

func ParseMessage(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

type JoinRoomPayload struct {
	RoomID string `json:"roomId"`
}

type LeaveRoomPayload struct {
	RoomID string `json:"roomId"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ConnectedPayload struct {
	UserID     string `json:"userId"`
	InstanceID string `json:"instanceId"`
}

type RoomJoinedPayload struct {
	RoomID      string `json:"roomId"`
	MemberCount int    `json:"memberCount"`
}

type RoomLeftPayload struct {
	RoomID string `json:"roomId"`
}

type PresenceUpdatePayload struct {
	UserID   string `json:"userId"`
	Username string `json:"username,omitempty"`
	Status   string `json:"status"` // "online" or "offline"
	RoomID   string `json:"roomId,omitempty"`
}

func NewErrorMessage(code, message string, requestID string) (*Message, error) {
	return NewMessageWithRequestID(MsgError, ErrorPayload{
		Code:    code,
		Message: message,
	}, requestID)
}

func NewPongMessage() (*Message, error) {
	return NewMessage(MsgPong, nil)
}
