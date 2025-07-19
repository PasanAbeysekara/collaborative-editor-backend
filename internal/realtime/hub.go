package realtime

import (
	"log"
	"time"
)

const (
	saveDebounceDuration = 3 * time.Second
)

type Hub struct {
	documentID string

	clients map[string]*Client

	broadcast chan []byte

	register chan *Client

	unregister chan *Client

	manager *Manager

	content string

	saveTimer *time.Timer
}

func newHub(docID string, initialContent string, m *Manager) *Hub {
	return &Hub{
		documentID: docID,
		content:    initialContent,
		clients:    make(map[string]*Client),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		manager:    m,
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.ID] = client
			log.Printf("Client %s registered to hub for document %s", client.ID, h.documentID)

			select {
			case client.send <- []byte(h.content):
				log.Printf("Sent initial content to client %s", client.ID)
			default:
				log.Printf("Failed to send initial content to client %s, unregistering.", client.ID)
				delete(h.clients, client.ID)
				close(client.send)
			}

		case client := <-h.unregister:
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.send)
				log.Printf("Client %s unregistered from hub for document %s", client.ID, h.documentID)

				if len(h.clients) == 0 {
					h.manager.removeHub(h)
					log.Printf("Hub for document %s is empty, removing it.", h.documentID)
					return
				}
			}

		case message := <-h.broadcast:
			h.content = string(message)

			for _, client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client.ID)
				}
			}

			if h.saveTimer != nil {
				h.saveTimer.Stop()
			}

			h.saveTimer = time.AfterFunc(saveDebounceDuration, func() {
				log.Printf("Debounce timer fired. Saving document %s to DB.", h.documentID)
				err := h.manager.Store.UpdateDocumentContent(h.documentID, h.content)
				if err != nil {
					log.Printf("ERROR: Failed to save document %s: %v", h.documentID, err)
				}
			})
		}
	}
}
