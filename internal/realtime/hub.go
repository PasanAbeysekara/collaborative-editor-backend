package realtime

import "log"

// Hub maintains the set of active clients for a single document
// and broadcasts messages to them.
type Hub struct {
	// The document this hub is for.
	documentID string

	// Registered clients. Maps a client pointer to a boolean.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Manager to notify when this hub is empty.
	manager *Manager
}

func newHub(docID string, m *Manager) *Hub {
	return &Hub{
		documentID: docID,
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		manager:    m,
	}
}

// run is the central loop for the hub. It must be run as a goroutine.
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("Client registered to hub for document %s", h.documentID)

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client unregistered from hub for document %s", h.documentID)

				// If the hub is now empty, tell the manager to remove it.
				if len(h.clients) == 0 {
					h.manager.removeHub(h)
					log.Printf("Hub for document %s is empty, removing it.", h.documentID)
					return // Stop the goroutine for this hub.
				}
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
					// Message sent successfully
				default:
					// If the send channel is blocked, the client is lagging.
					// We unregister them to avoid blocking the hub.
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
