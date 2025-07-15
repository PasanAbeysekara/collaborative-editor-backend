package realtime

import (
	"log"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for now. In production, you'd want to restrict this.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Manager is the top-level controller for all real-time hubs.
type Manager struct {
	hubs map[string]*Hub
	mu   sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		hubs: make(map[string]*Hub),
	}
}

// getOrCreateHub finds a hub for a given documentID or creates a new one.
func (m *Manager) getOrCreateHub(documentID string) *Hub {
	m.mu.Lock()
	defer m.mu.Unlock()

	if hub, ok := m.hubs[documentID]; ok {
		return hub
	}

	hub := newHub(documentID, m)
	m.hubs[documentID] = hub
	go hub.run() // Start the hub's processing loop
	log.Printf("Created a new hub for document %s", documentID)
	return hub
}

// removeHub is called by a hub when it's empty to remove itself from the manager.
func (m *Manager) removeHub(hub *Hub) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.hubs[hub.documentID]; ok {
		delete(m.hubs, hub.documentID)
	}
}

// ServeWS is the HTTP handler for WebSocket connections.
func (m *Manager) ServeWS(w http.ResponseWriter, r *http.Request) {
	// In a real app, you must authenticate the user and authorize access
	// to the document *before* upgrading the connection.
	// We'll skip that for now for simplicity.

	documentID := chi.URLParam(r, "documentID")
	if documentID == "" {
		http.Error(w, "Document ID is required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade connection:", err)
		return
	}

	hub := m.getOrCreateHub(documentID)
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	// Start the read and write pumps in separate goroutines.
	go client.writePump()
	go client.readPump()

	log.Printf("Client connected to hub for document %s", documentID)
}
