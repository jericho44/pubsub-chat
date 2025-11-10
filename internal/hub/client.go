package hub

import (
	"encoding/json"
	"pubsub-chat/internal/logger"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 5120 // <-- tambahkan baris ini
)

// Client represents a single websocket connection.
type Client struct {
	name string
	rm   *Room
	conn *websocket.Conn
	send chan []byte
	lg   *logger.Logger
}

func NewClient(name string, rm *Room, c *websocket.Conn, lg *logger.Logger) *Client {
	return &Client{name: name, rm: rm, conn: c, send: make(chan []byte, 256), lg: lg}
}

func (c *Client) readPump() {
	defer func() {
		c.rm.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		// Expect JSON: {type:"chat", room:"...", sender:"...", text:"..."}
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			c.lg.Errorw("ws.unmarshal", "err", err)
			continue
		}
		if msg.Room == "" {
			msg.Room = c.rm.name
		}
		if msg.Sender == "" {
			msg.Sender = c.name
		}
		encoded, _ := json.Marshal(msg)
		c.rm.broadcast <- encoded
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// channel closed
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			if _, err := w.Write(message); err != nil {
				return
			}
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
