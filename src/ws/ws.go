package ws

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	mu    sync.Mutex
	conns map[*websocket.Conn]bool
}

var hub = &Hub{
	conns: make(map[*websocket.Conn]bool),
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	hub.add(conn)

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			hub.remove(conn)
			return
		}
	}
}

func NotifyUpdate() {
	hub.broadcast([]byte("update"))
}

func (h *Hub) add(c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conns[c] = true
}

func (h *Hub) remove(c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.conns, c)
	c.Close()
}

func (h *Hub) broadcast(msg []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for c := range h.conns {
		if err := c.WriteMessage(websocket.TextMessage, msg); err != nil {
			delete(h.conns, c)
			c.Close()
		}
	}
}
