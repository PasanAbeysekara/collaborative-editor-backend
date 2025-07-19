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
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Manager struct {
	hubs map[string]*Hub
	mu   sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		hubs: make(map[string]*Hub),
	}
}

func (m *Manager) getOrCreateHub(documentID string) *Hub {
	m.mu.Lock()
	defer m.mu.Unlock()

	if hub, ok := m.hubs[documentID]; ok {
		return hub
	}

	hub := newHub(documentID, m)
	m.hubs[documentID] = hub
	go hub.run()
	log.Printf("Created a new hub for document %s", documentID)
	return hub
}

func (m *Manager) removeHub(hub *Hub) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.hubs[hub.documentID]; ok {
		delete(m.hubs, hub.documentID)
	}
}

func (m *Manager) ServeWS(w http.ResponseWriter, r *http.Request) {
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

	go client.writePump()
	go client.readPump()

	log.Printf("Client connected to hub for document %s", documentID)
}
