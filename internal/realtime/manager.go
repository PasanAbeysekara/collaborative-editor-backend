package realtime

import (
	"log"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/pasanAbeysekara/collaborative-editor/internal/auth"
	"github.com/pasanAbeysekara/collaborative-editor/internal/storage"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Manager struct {
	hubs map[string]*Hub
	mu   sync.RWMutex

	Store storage.Store
}

func NewManager(store storage.Store) *Manager {
	return &Manager{
		hubs:  make(map[string]*Hub),
		Store: store,
	}
}

func (m *Manager) getOrCreateHub(documentID string) (*Hub, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if hub, ok := m.hubs[documentID]; ok {
		return hub, nil
	}

	doc, err := m.Store.GetDocument(documentID)
	if err != nil {
		log.Printf("Failed to get document %s from store: %v", documentID, err)
		return nil, err
	}

	hub := newHub(documentID, doc.Content, m)
	m.hubs[documentID] = hub
	go hub.run()
	log.Printf("Created a new hub for document %s", documentID)
	return hub, nil
}

func (m *Manager) removeHub(hub *Hub) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.hubs[hub.documentID]; ok {
		delete(m.hubs, hub.documentID)
	}
}

func (m *Manager) ServeWS(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized: User ID not found in token", http.StatusUnauthorized)
		return
	}

	documentID := chi.URLParam(r, "documentID")
	if documentID == "" {
		http.Error(w, "Bad Request: Document ID is required", http.StatusBadRequest)
		return
	}

	hasPermission, err := m.Store.CheckDocumentPermission(documentID, userID)
	if err != nil {
		log.Printf("Error checking document permission for doc %s, user %s: %v", documentID, userID, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if !hasPermission {
		log.Printf("Forbidden: User %s has no permission for document %s", userID, documentID)
		http.Error(w, "Forbidden: You do not have access to this document", http.StatusForbidden)
		return
	}

	log.Printf("Authorization successful for user %s on document %s. Upgrading connection...", userID, documentID)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade connection:", err)
		return
	}

	hub, err := m.getOrCreateHub(documentID)
	if err != nil {
		http.Error(w, "Document not found or internal error", http.StatusNotFound)
		return
	}

	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()

	log.Printf("Client for user %s connected to hub for document %s", userID, documentID)
}
