package socket

import (
    "context"
    "fmt"
    "net/http"
    "os"
    "path/filepath"
    "sync"
    "time"

    "github.com/gorilla/websocket"
    "github.com/rs/zerolog/log"
)

type APIKeyDetails struct {
    APIKey    string `json:"api_key"`
    Username  string `json:"username"`
    CreatedAt string `json:"created_at"`
}

type Hub struct {
    clients    map[*Client]bool
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
    mutex      sync.RWMutex
}

type Client struct {
    hub     *Hub
    conn    *websocket.Conn
    send    chan []byte
    apiKey  string
    ctx     context.Context
    cancel  context.CancelFunc
}

type WebSocketHandler struct {
    validAPIKeys map[string]APIKeyDetails
    hub          *Hub
    upgrader     websocket.Upgrader
}

func NewWebSocketHandler() *WebSocketHandler {
    validKeys := map[string]APIKeyDetails{
        "eofqjvwilceiecjwqludjtshrcnfcduo": {
            APIKey:    "eofqjvwilceiecjwqludjtshrcnfcduo",
            Username:  "enigma",
            CreatedAt: "2024-08-07T17:20:31.087Z",
        },
        "lursymyslygmjrmvuvzieekakkogjo": {
            APIKey:    "lursymyslygmjrmvuvzieekakkogjo",
            Username:  "martin",
            CreatedAt: "2024-08-07T17:20:42.236Z",
        },
    }

    hub := &Hub{
        clients:    make(map[*Client]bool),
        broadcast:  make(chan []byte),
        register:   make(chan *Client),
        unregister: make(chan *Client),
    }

    upgrader := websocket.Upgrader{
        ReadBufferSize:  1024,
        WriteBufferSize: 1024,
        CheckOrigin: func(r *http.Request) bool {
            return true
        },
        HandshakeTimeout: 45 * time.Second,
    }

    handler := &WebSocketHandler{
        validAPIKeys: validKeys,
        hub:          hub,
        upgrader:     upgrader,
    }

    // Start the hub
    go handler.hub.run()

    return handler
}

func (h *Hub) run() {
    for {
        select {
        case client := <-h.register:
            h.mutex.Lock()
            h.clients[client] = true
            h.mutex.Unlock()
            log.Info().Str("api_key", client.apiKey).Msg("Client connected")

        case client := <-h.unregister:
            h.mutex.Lock()
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
            }
            h.mutex.Unlock()
            log.Info().Str("api_key", client.apiKey).Msg("Client disconnected")

        case message := <-h.broadcast:
            h.mutex.RLock()
            for client := range h.clients {
                select {
                case client.send <- message:
                default:
                    delete(h.clients, client)
                    close(client.send)
                }
            }
            h.mutex.RUnlock()
        }
    }
}

func (wsh *WebSocketHandler) isValidAPIKey(key string) bool {
    _, exists := wsh.validAPIKeys[key]
    return exists
}

func (wsh *WebSocketHandler) getLogFilePath(apiKey string) string {
    return filepath.Join("logs", fmt.Sprintf("%s.log", apiKey))
}

func (wsh *WebSocketHandler) createLogsDir() error {
    return os.MkdirAll("logs", 0755)
}

func (wsh *WebSocketHandler) logMessage(apiKey, messageType, message string) error {
    filePath := wsh.getLogFilePath(apiKey)
    
    file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return fmt.Errorf("failed to open log file: %w", err)
    }
    defer file.Close()

    logEntry := fmt.Sprintf("%s (%s): %s\n", messageType, time.Now().Format(time.RFC3339), message)
    if _, err := file.WriteString(logEntry); err != nil {
        return fmt.Errorf("failed to write to log file: %w", err)
    }

    return nil
}

func (wsh *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    apiKey := r.URL.Query().Get("apiKey")
    if apiKey == "" {
        http.Error(w, "API Key is required", http.StatusBadRequest)
        return
    }

    if !wsh.isValidAPIKey(apiKey) {
        http.Error(w, "Invalid API Key", http.StatusUnauthorized)
        return
    }

    if err := wsh.createLogsDir(); err != nil {
        log.Error().Err(err).Msg("Failed to create logs directory")
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    conn, err := wsh.upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Error().Err(err).Msg("Failed to upgrade connection")
        return
    }

    ctx, cancel := context.WithCancel(context.Background())
    client := &Client{
        hub:    wsh.hub,
        conn:   conn,
        send:   make(chan []byte, 256),
        apiKey: apiKey,
        ctx:    ctx,
        cancel: cancel,
    }

    wsh.hub.register <- client

    go client.writePump(wsh)
    go client.readPump(wsh)
}

const (
    writeWait      = 10 * time.Second
    pongWait       = 60 * time.Second
    pingPeriod     = (pongWait * 9) / 10
    maxMessageSize = 512
)

func (c *Client) readPump(wsh *WebSocketHandler) {
    defer func() {
        c.hub.unregister <- c
        c.conn.Close()
        c.cancel()
    }()

    c.conn.SetReadLimit(maxMessageSize)
    c.conn.SetReadDeadline(time.Now().Add(pongWait))
    c.conn.SetPongHandler(func(string) error {
        c.conn.SetReadDeadline(time.Now().Add(pongWait))
        return nil
    })

    for {
        select {
        case <-c.ctx.Done():
            return
        default:
            messageType, message, err := c.conn.ReadMessage()
            if err != nil {
                if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                    log.Error().Err(err).Str("api_key", c.apiKey).Msg("WebSocket error")
                }
                return
            }

            if err := wsh.logMessage(c.apiKey, "RECEIVED", string(message)); err != nil {
                log.Error().Err(err).Str("api_key", c.apiKey).Msg("Failed to log received message")
            }

            switch messageType {
            case websocket.TextMessage, websocket.BinaryMessage:
                select {
                case c.hub.broadcast <- message:
                case <-c.ctx.Done():
                    return
                }
            }
        }
    }
}

func (c *Client) writePump(wsh *WebSocketHandler) {
    ticker := time.NewTicker(pingPeriod)
    defer func() {
        ticker.Stop()
        c.conn.Close()
        c.cancel()
    }()

    for {
        select {
        case message, ok := <-c.send:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if !ok {
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
                log.Error().Err(err).Str("api_key", c.apiKey).Msg("Failed to write message")
                return
            }
            if err := wsh.logMessage(c.apiKey, "SENT", string(message)); err != nil {
                log.Error().Err(err).Str("api_key", c.apiKey).Msg("Failed to log sent message")
            }

        case <-ticker.C:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                log.Error().Err(err).Str("api_key", c.apiKey).Msg("Failed to send ping")
                return
            }

        case <-c.ctx.Done():
            return
        }
    }
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    handler := NewWebSocketHandler()
    handler.HandleWebSocket(w, r)
}
