package realtime

import (
	"encoding/json"
	"sync"

	"github.com/gofiber/websocket/v2"
)

type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[*websocket.Conn]struct{} // room = boardIdHex
}

var H = NewHub()

func NewHub() *Hub {
	return &Hub{rooms: make(map[string]map[*websocket.Conn]struct{})}
}

func (h *Hub) Join(room string, c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[room] == nil {
		h.rooms[room] = make(map[*websocket.Conn]struct{})
	}
	h.rooms[room][c] = struct{}{}
}

func (h *Hub) Leave(room string, c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if cs, ok := h.rooms[room]; ok {
		delete(cs, c)
		if len(cs) == 0 {
			delete(h.rooms, room)
		}
	}
}

func (h *Hub) Broadcast(room string, evt Event) {
	h.mu.RLock()
	conns := h.rooms[room]
	h.mu.RUnlock()

	if len(conns) == 0 {
		return
	}
	b, _ := json.Marshal(evt)
	for c := range conns {
		_ = c.WriteMessage(websocket.TextMessage, b)
	}
}
