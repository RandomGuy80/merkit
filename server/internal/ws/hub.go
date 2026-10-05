package ws

import (
	"encoding/json"
	"sync"

	"github.com/gofiber/contrib/websocket"
)

type Message struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

type client struct {
	conn   *websocket.Conn
	userID string
	send   chan []byte
}

type Hub struct {
	mu      sync.RWMutex
	clients map[string][]*client
}

func NewHub() *Hub {
	return &Hub{clients: make(map[string][]*client)}
}

func (h *Hub) Register(userID string, conn *websocket.Conn) *client {
	c := &client{conn: conn, userID: userID, send: make(chan []byte, 32)}
	h.mu.Lock()
	h.clients[userID] = append(h.clients[userID], c)
	h.mu.Unlock()
	return c
}

func (h *Hub) Unregister(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	list := h.clients[c.userID]
	for i, cl := range list {
		if cl == c {
			h.clients[c.userID] = append(list[:i], list[i+1:]...)
			break
		}
	}
	close(c.send)
}

func (h *Hub) Send(userID string, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	h.mu.RLock()
	clients := h.clients[userID]
	h.mu.RUnlock()
	for _, c := range clients {
		select {
		case c.send <- data:
		default:
		}
	}
}

func (h *Hub) RunClient(c *client) {
	go func() {
		for msg := range c.send {
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		}
	}()
	// drain incoming (keep-alive)
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
	h.Unregister(c)
}
