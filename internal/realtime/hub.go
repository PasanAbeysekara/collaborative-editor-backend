package realtime

import (
	"log"
	"time"
)

const (
	saveDebounceDuration = 3 * time.Second
)

type OpPayload struct {
	SourceClient *Client
	Op           *Operation
}

type Hub struct {
	documentID string

	clients map[string]*Client

	incomingOps chan *OpPayload

	register chan *Client

	unregister chan *Client

	manager *Manager

	content string
	// A timer to debounce database save operations.
	saveTimer *time.Timer

	version int
}

func newHub(docID string, initialContent string, initialVersion int, m *Manager) *Hub {
	return &Hub{
		documentID:  docID,
		content:     initialContent,
		version:     initialVersion,
		clients:     make(map[string]*Client),
		incomingOps: make(chan *OpPayload),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		manager:     m,
	}
}

func (h *Hub) applyOperation(op *Operation) {
	if op.Type == OpInsert {
		h.content = h.content[:op.Pos] + op.Text + h.content[op.Pos:]
	} else if op.Type == OpDelete {
		h.content = h.content[:op.Pos] + h.content[op.Pos+op.Len:]
	}
	h.version++
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

		case payload := <-h.incomingOps:
			op := payload.Op

			if op.Version != h.version {
				log.Printf("Conflict on doc %s: op version %d, server version %d. Op rejected.",
					h.documentID, op.Version, h.version)
				continue
			}

			h.applyOperation(op)

			op.Version = h.version

			log.Printf("Broadcasting op (v%d) to %d clients for doc %s", op.Version, len(h.clients), h.documentID)
			for _, client := range h.clients {
				if client.ID == payload.SourceClient.ID {
					continue
				}
				select {
				case client.sendOp <- op:
				default:
					close(client.sendOp)
					delete(h.clients, client.ID)
				}
			}

			if h.saveTimer != nil {
				h.saveTimer.Stop()
			}
			h.saveTimer = time.AfterFunc(saveDebounceDuration, func() {
				log.Printf("Debounce timer fired. Saving document %s (v%d) to DB.", h.documentID, h.version)
				err := h.manager.Store.UpdateDocument(h.documentID, h.content, h.version)
				if err != nil {
					log.Printf("ERROR: Failed to save document %s: %v", h.documentID, err)
				}
			})
		}
	}
}
