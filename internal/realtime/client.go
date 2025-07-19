package realtime

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

type Client struct {
	ID   string
	hub  *Hub
	conn *websocket.Conn
	send chan *ServerMessage
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error for client %s: %v", c.ID, err)
			break
		}

		var op Operation
		if err := json.Unmarshal(message, &op); err != nil {
			log.Printf("Failed to unmarshal operation from client %s: %v", c.ID, err)
			continue
		}

		payload := &OpPayload{
			SourceClient: c,
			Op:           &op,
		}
		c.hub.incomingOps <- payload
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
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			// Marshal the entire message struct to JSON.
			if err := c.conn.WriteJSON(message); err != nil {
				log.Printf("Failed to write JSON message to client %s: %v", c.ID, err)
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
