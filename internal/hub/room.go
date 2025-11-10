package hub

import "sync"

// Room is a topic holding many clients.
type Room struct {
	name       string
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewRoom(name string) *Room {
	return &Room{
		name:       name,
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (r *Room) Run() {
	for {
		select {
		case c := <-r.register:
			r.mu.Lock()
			r.clients[c] = true
			r.mu.Unlock()
		case c := <-r.unregister:
			r.mu.Lock()
			if _, ok := r.clients[c]; ok {
				delete(r.clients, c)
				close(c.send)
			}
			r.mu.Unlock()
		case msg := <-r.broadcast:
			r.mu.RLock()
			for c := range r.clients {
				select {
				case c.send <- msg:
				default:
					// Slow consumer: drop & disconnect
					delete(r.clients, c)
					close(c.send)
				}
			}
			r.mu.RUnlock()
		}
	}
}
