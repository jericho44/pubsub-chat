package hub

import (
	"encoding/json"
	"net/http"
	"pubsub-chat/internal/logger"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type   string `json:"type"` // "join" | "leave" | "chat"
	Room   string `json:"room"`
	Sender string `json:"sender"`
	Text   string `json:"text"`
}

// Hub coordinates rooms (topics) and broadcasts.
type Hub struct {
	rooms map[string]*Room
	mu    sync.RWMutex
	upgr  websocket.Upgrader
	quit  chan struct{}
}

func New(lg *logger.Logger) *Hub {
	return &Hub{
		rooms: make(map[string]*Room),
		upgr: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
		quit: make(chan struct{}),
	}
}

func (h *Hub) getOrCreateRoom(name string) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()
	if rm, ok := h.rooms[name]; ok {
		return rm
	}
	rm := NewRoom(name)
	h.rooms[name] = rm
	go rm.Run()
	return rm
}

func (h *Hub) Run() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			h.gcRooms()
		case <-h.quit:
			return
		}
	}
}

func (h *Hub) Stop() { // panggil saat shutdown
	close(h.quit)
}

func (h *Hub) gcRooms() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for name, rm := range h.rooms {
		rm.mu.RLock()
		empty := len(rm.clients) == 0
		rm.mu.RUnlock()
		if empty {
			delete(h.rooms, name)
		}
	}
}

func (h *Hub) ServeWS(lg *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		room := r.URL.Query().Get("room")
		if room == "" {
			room = "general"
		}
		name := r.URL.Query().Get("name")
		if name == "" {
			name = "anon"
		}

		conn, err := h.upgr.Upgrade(w, r, nil)
		if err != nil {
			lg.Errorw("ws.upgrade", "err", err)
			return
		}

		rm := h.getOrCreateRoom(room)
		cl := NewClient(name, rm, conn, lg)
		rm.register <- cl

		// Send a join message system-broadcast
		joinMsg, _ := json.Marshal(Message{Type: "join", Room: room, Sender: name, Text: name + " joined"})
		rm.broadcast <- joinMsg

		go cl.writePump()
		cl.readPump()
	}
}
