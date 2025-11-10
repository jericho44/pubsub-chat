package hub

import "testing"

func TestRoomLifecycle(t *testing.T) {
	r := NewRoom("test")
	go r.Run()
	// Register/unregister sanity
	c := &Client{rm: r, send: make(chan []byte, 1)}
	r.register <- c
	r.unregister <- c
}
