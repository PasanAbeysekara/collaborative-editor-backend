package realtime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/redis/go-redis/v9"
)

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

func invertOperation(op *Operation) *Operation {
	if op.Type == OpInsert {
		return &Operation{Type: OpDelete, Pos: op.Pos, Len: len(op.Text)}
	} else if op.Type == OpDelete {
		return &Operation{Type: OpInsert, Pos: op.Pos, Text: op.Text}
	}
	return nil
}

// sendErrorToClient is a helper to send a targeted error message to a single client.
func (h *Hub) sendErrorToClient(client *Client, errorMsg string) {
	errMessage := &ServerMessage{
		Type:  MsgError,
		Error: errorMsg,
	}
	select {
	case client.send <- errMessage:
	default:
		log.Printf("Failed to send error message to client %s, channel full.", client.ID)
	}
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
				Version: h.version,
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
					log.Printf("Hub for doc %s is now empty. Saving final state via document-service.", h.documentID)

					jsonData, err := json.Marshal(map[string]interface{}{
						"content": h.content,
						"version": h.version,
					})
					if err != nil {
						log.Printf("CRITICAL: Failed to marshal final state for doc %s: %v", h.documentID, err)
					} else {
						url := fmt.Sprintf("%s/documents/%s", h.manager.documentServiceURL, h.documentID)
						req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
						if err != nil {
							log.Printf("CRITICAL: Failed to create save request for doc %s: %v", h.documentID, err)
						} else {
							req.Header.Set("Content-Type", "application/json")
							resp, err := http.DefaultClient.Do(req)
							if err != nil || resp.StatusCode != http.StatusOK {
								status := 0
								if resp != nil {
									status = resp.StatusCode
								}
								log.Printf("CRITICAL: Failed to save final state for doc %s via API: %v (status: %d)", h.documentID, err, status)
							}
						}
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
			client := payload.SourceClient // Get the originating client

			// --- 1. HANDLE VERSION MISMATCH ---
			// Don't check version for 'undo' operations.
			if op.Type != OpUndo && op.Version != h.version {
				log.Printf("Version mismatch for client %s on doc %s. Client: v%d, Server: v%d. Sending sync state.",
					client.ID, h.documentID, op.Version, h.version)

				// Send the correct, latest state back to the out-of-sync client.
				syncMsg := &ServerMessage{
					Type:    MsgOutOfSync,
					Content: h.content,
					Version: h.version,
				}
				select {
				case client.send <- syncMsg:
				default:
					log.Printf("Failed to send sync state to client %s, channel full.", client.ID)
				}
				continue // Stop processing this outdated operation.
			}

			// --- 2. HANDLE UNDO OPERATION ---
			if op.Type == OpUndo {
				lastOpData, err := h.manager.Cache.PopOperation(context.Background(), h.documentID)
				if err != nil {
					if err == redis.Nil {
						log.Printf("Undo requested for doc %s, but no operations to undo.", h.documentID)
						h.sendErrorToClient(client, "No operations to undo.")
					} else {
						log.Printf("ERROR: Failed to pop operation for undo on doc %s: %v", h.documentID, err)
						h.sendErrorToClient(client, "Server error while processing undo.")
					}
					continue
				}
				var lastOp Operation
				if err := json.Unmarshal(lastOpData, &lastOp); err != nil {
					log.Printf("ERROR: Failed to unmarshal last op for undo: %v", err)
					h.sendErrorToClient(client, "Server error: cannot process last operation.")
					continue
				}
				invertedOp := invertOperation(&lastOp)
				if invertedOp == nil {
					log.Printf("WARN: Could not invert operation of type %s", lastOp.Type)
					h.sendErrorToClient(client, "Cannot undo the last action.")
					continue
				}
				h.applyOperation(invertedOp)
				invertedOp.Version = h.version
				if err := h.manager.Cache.SetDocumentState(context.Background(), h.documentID, h.content, h.version); err != nil {
					log.Printf("WARN: Failed to save state to cache for doc %s after undo: %v", h.documentID, err)
				}
				opMsg := &ServerMessage{Type: MsgOperation, Op: invertedOp}
				// Broadcast undo to all clients so everyone's state is updated.
				for _, otherClient := range h.clients {
					select {
					case otherClient.send <- opMsg:
					default:
						close(otherClient.send)
						delete(h.clients, otherClient.ID)
					}
				}
				continue // Finish processing the undo operation.
			}

			// --- 3. VALIDATE OTHER EDGE CASES ---
			switch op.Type {
			case OpInsert:
				if op.Pos < 0 || op.Pos > len(h.content) {
					h.sendErrorToClient(client, fmt.Sprintf("Invalid position for insert: %d. Max position is %d.", op.Pos, len(h.content)))
					continue
				}
			case OpDelete:
				if op.Pos < 0 || op.Pos+op.Len > len(h.content) {
					h.sendErrorToClient(client, fmt.Sprintf("Invalid range for delete: pos %d, len %d. Content length is %d.", op.Pos, op.Len, len(h.content)))
					continue
				}
			default:
				h.sendErrorToClient(client, fmt.Sprintf("Unsupported operation type: '%s'.", op.Type))
				continue
			}

			// --- 4. IF ALL CHECKS PASS, PROCESS THE OPERATION ---
			if op.Type == OpDelete {
				if op.Pos+op.Len <= len(h.content) {
					op.Text = h.content[op.Pos : op.Pos+op.Len]
				}
			}

			h.applyOperation(op)

			opBytes, err := json.Marshal(op)
			if err == nil {
				h.manager.Cache.PushOperation(context.Background(), h.documentID, opBytes)
			}

			if err := h.manager.Cache.SetDocumentState(context.Background(), h.documentID, h.content, h.version); err != nil {
				log.Printf("WARN: Failed to save state to cache for doc %s: %v", h.documentID, err)
			}

			op.Version = h.version
			opMsg := &ServerMessage{Type: MsgOperation, Op: op}
			for _, otherClient := range h.clients {
				// Broadcast the successful operation to all *other* clients.
				if otherClient.ID == client.ID {
					continue
				}
				select {
				case otherClient.send <- opMsg:
				default:
					close(otherClient.send)
					delete(h.clients, otherClient.ID)
				}
			}
		}
	}
}