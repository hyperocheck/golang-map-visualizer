package ws

import (
	"net/http"
	"sync"
	"log"

	"github.com/gorilla/websocket"
)

/*
	Hub хранит ВСЕ websocket-клиенты
*/
type Hub struct {
	mu    sync.Mutex
	conns map[*websocket.Conn]bool
}

/*
	Единственный hub на всё приложение
*/
var hub = &Hub{
	conns: make(map[*websocket.Conn]bool),
}

/*
	WebSocket upgrader
*/
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // dev
	},
}

/*
	HTTP handler для /ws
*/
func Handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	hub.add(conn)

	for {
		// читаем, чтобы ловить disconnect
		if _, _, err := conn.ReadMessage(); err != nil {
			hub.remove(conn)
			return
		}
	}
}

/*
	ПУБЛИЧНАЯ функция
	Вызывается ИЗ ЛЮБОГО ПАКЕТА
*/
func NotifyUpdate() {
	log.Println("[ws] notify update")
	hub.broadcast([]byte("update"))
}

/*
	----- приватная логика ниже -----
*/

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

