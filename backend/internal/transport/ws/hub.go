package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/szabolcs/cms/internal/infrastructure"
	"github.com/szabolcs/cms/internal/service"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Hub manages WebSocket connections and broadcasts events via Redis pub/sub.
type Hub struct {
	rdb    *redis.Client
	events service.EventLister
	logger *slog.Logger

	register   chan *websocket.Conn
	unregister chan *websocket.Conn

	mu      sync.RWMutex
	clients map[*websocket.Conn]struct{}
}

// NewHub creates a new WebSocket Hub.
func NewHub(rdb *redis.Client, events service.EventLister, logger *slog.Logger) *Hub {
	return &Hub{
		rdb:        rdb,
		events:     events,
		logger:     logger,
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		clients:    make(map[*websocket.Conn]struct{}),
	}
}

// Run starts the hub's register/unregister loop and Redis subscriber.
func (h *Hub) Run(ctx context.Context) error {
	// Start Redis subscription in a separate goroutine.
	go h.subscribeRedis(ctx)

	for {
		select {
		case <-ctx.Done():
			h.mu.Lock()
			for conn := range h.clients {
				conn.Close()
			}
			h.clients = make(map[*websocket.Conn]struct{})
			h.mu.Unlock()
			return nil
		case conn := <-h.register:
			h.mu.Lock()
			h.clients[conn] = struct{}{}
			h.mu.Unlock()
			h.logger.Info("ws: client connected", "clients", h.clientCount())
			// Send last 10 events as initial state.
			go h.sendInitialState(ctx, conn)
		case conn := <-h.unregister:
			h.mu.Lock()
			delete(h.clients, conn)
			h.mu.Unlock()
			conn.Close()
			h.logger.Info("ws: client disconnected", "clients", h.clientCount())
		}
	}
}

// HandleWS is the gin handler for WebSocket upgrade.
func (h *Hub) HandleWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("ws: upgrade failed", "error", err)
		return
	}

	h.register <- conn

	// Read loop to detect disconnection.
	go func() {
		defer func() { h.unregister <- conn }()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()
}

func (h *Hub) subscribeRedis(ctx context.Context) {
	sub := h.rdb.Subscribe(ctx, infrastructure.RedisChannelValidationEvents)
	defer sub.Close()

	ch := sub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			h.broadcast([]byte(msg.Payload))
		}
	}
}

func (h *Hub) broadcast(data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	msg := struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}{
		Type: "event",
		Data: data,
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		h.logger.Error("ws: failed to marshal broadcast", "error", err)
		return
	}

	for conn := range h.clients {
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			h.logger.Error("ws: write failed", "error", err)
			go func(c *websocket.Conn) { h.unregister <- c }(conn)
		}
	}
}

func (h *Hub) sendInitialState(ctx context.Context, conn *websocket.Conn) {
	events, err := h.events.GetRecentEvents(ctx, 10)
	if err != nil {
		h.logger.Error("ws: failed to get initial events", "error", err)
		return
	}

	msg := struct {
		Type string `json:"type"`
		Data any    `json:"data"`
	}{
		Type: "initial",
		Data: events,
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		h.logger.Error("ws: failed to marshal initial state", "error", err)
		return
	}

	h.mu.RLock()
	_, exists := h.clients[conn]
	h.mu.RUnlock()
	if !exists {
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
		h.logger.Error("ws: failed to send initial state", "error", err)
	}
}

func (h *Hub) clientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
