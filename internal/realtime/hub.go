package realtime

import (
	"context"
	"encoding/json"
	"log"

	"github.com/redis/go-redis/v9"
)

// OpPayload is a wrapper for an operation that includes its source client.
type OpPayload struct {
	SourceClient *Client
	Op           *Operation
}

type Hub struct {
	documentID  string
	clients     map[string]*Client
	incomingOps chan *OpPayload
	register    chan *Client
	unregister  chan *Client
	manager     *Manager
	content     string
	version     int
}

func newHub(docID, initialContent string, initialVersion int, m *Manager) *Hub {
	return &Hub{
		documentID:  docID,
		content:     initialContent,
		version:     initialVersion,
		manager:     m,
		clients:     make(map[string]*Client),
		incomingOps: make(chan *OpPayload),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
	}
}

// applyOperation applies an op to the hub's content.
func (h *Hub) applyOperation(op *Operation) {
	if op.Type == OpInsert {
		if op.Pos > len(h.content) {
			op.Pos = len(h.content)
		}
		h.content = h.content[:op.Pos] + op.Text + h.content[op.Pos:]
	} else if op.Type == OpDelete {
		if op.Pos+op.Len > len(h.content) {
			op.Len = len(h.content) - op.Pos
		}
		if op.Pos < len(h.content) {
			h.content = h.content[:op.Pos] + h.content[op.Pos+op.Len:]
		}
	}
	h.version++
}

// invertOperation creates the opposite of a given operation for undo purposes.
func invertOperation(op *Operation) *Operation {
	if op.Type == OpInsert {
		return &Operation{
			Type: OpDelete,
			Pos:  op.Pos,
			Len:  len(op.Text),
		}
	} else if op.Type == OpDelete {
		return &Operation{
			Type: OpInsert,
			Pos:  op.Pos,
			Text: op.Text,
		}
	}
	return nil
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.ID] = client
			log.Printf("Client %s registered to hub for document %s", client.ID, h.documentID)
			initialStateMsg := &ServerMessage{
				Type:    MsgInitialState,
				Content: h.content,
			}

			select {
			case client.send <- initialStateMsg:
				log.Printf("Sent initial state to client %s", client.ID)
			default:
				log.Printf("Failed to send initial state to client %s, unregistering.", client.ID)
				delete(h.clients, client.ID)
				close(client.send)
			}

		case client := <-h.unregister:
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.send)
				log.Printf("Client %s unregistered from hub for document %s", client.ID, h.documentID)

				if len(h.clients) == 0 {
					log.Printf("Hub for doc %s is now empty. Saving final state to PostgreSQL.", h.documentID)
					if err := h.manager.Store.UpdateDocument(h.documentID, h.content, h.version); err != nil {
						log.Printf("CRITICAL: Failed to save final state for doc %s to DB: %v", h.documentID, err)
					}
					if err := h.manager.Cache.ClearDocumentState(context.Background(), h.documentID); err != nil {
						log.Printf("WARN: Failed to clear cache for doc %s: %v", h.documentID, err)
					}
					h.manager.removeHub(h)
					log.Printf("Hub for document %s is shutting down.", h.documentID)
					return
				}
			}

		case payload := <-h.incomingOps:
			op := payload.Op

			// =========================================================
			// === THIS IS THE CRITICAL LOGIC THAT WAS MISSING =========
			// =========================================================
			if op.Type == OpUndo {
				lastOpData, err := h.manager.Cache.PopOperation(context.Background(), h.documentID)
				if err != nil {
					if err == redis.Nil {
						log.Printf("Undo requested for doc %s, but no operations to undo.", h.documentID)
					} else {
						log.Printf("ERROR: Failed to pop operation for undo on doc %s: %v", h.documentID, err)
					}
					continue
				}

				var lastOp Operation
				if err := json.Unmarshal(lastOpData, &lastOp); err != nil {
					log.Printf("ERROR: Failed to unmarshal last op for undo: %v", err)
					continue
				}

				invertedOp := invertOperation(&lastOp)
				if invertedOp == nil {
					log.Printf("WARN: Could not invert operation of type %s", lastOp.Type)
					continue
				}

				h.applyOperation(invertedOp)
				invertedOp.Version = h.version

				if err := h.manager.Cache.SetDocumentState(context.Background(), h.documentID, h.content, h.version); err != nil {
					log.Printf("WARN: Failed to save state to cache for doc %s after undo: %v", h.documentID, err)
				}

				opMsg := &ServerMessage{Type: MsgOperation, Op: invertedOp}
				for _, client := range h.clients {
					select {
					case client.send <- opMsg:
					default:
						close(client.send)
						delete(h.clients, client.ID)
					}
				}
				continue // Skip the normal operation handling
			}
			// === END OF UNDO LOGIC ===================================

			// --- NORMAL OPERATION LOGIC ---
			if op.Version != h.version {
				log.Printf("Conflict on doc %s: op version %d, server version %d. Op rejected.", h.documentID, op.Version, h.version)
				continue
			}

			// Capture deleted text *before* applying the operation.
			if op.Type == OpDelete {
				if op.Pos+op.Len <= len(h.content) {
					op.Text = h.content[op.Pos : op.Pos+op.Len]
				}
			}

			h.applyOperation(op)

			// Push the completed operation to the undo stack in Redis.
			opBytes, err := json.Marshal(op)
			if err == nil {
				h.manager.Cache.PushOperation(context.Background(), h.documentID, opBytes)
			}

			if err := h.manager.Cache.SetDocumentState(context.Background(), h.documentID, h.content, h.version); err != nil {
				log.Printf("WARN: Failed to save state to cache for doc %s: %v", h.documentID, err)
			}

			op.Version = h.version
			opMsg := &ServerMessage{Type: MsgOperation, Op: op}

			for _, client := range h.clients {
				if client.ID == payload.SourceClient.ID {
					continue
				}
				select {
				case client.send <- opMsg:
				default:
					close(client.send)
					delete(h.clients, client.ID)
				}
			}
		}
	}
}
