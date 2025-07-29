package realtime

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pasanAbeysekara/collaborative-editor/internal/auth"
	"github.com/pasanAbeysekara/collaborative-editor/internal/storage"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Manager struct {
	hubs  map[string]*Hub
	mu    sync.RWMutex
	Store storage.Store
	Cache storage.Cache
}

func NewManager(store storage.Store, cache storage.Cache) *Manager {
	return &Manager{
		hubs:  make(map[string]*Hub),
		Store: store,
		Cache: cache,
	}
}

func (m *Manager) getOrCreateHub(documentID string) (*Hub, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if hub, ok := m.hubs[documentID]; ok {
		return hub, nil
	}

	content, version, err := m.Cache.GetDocumentState(context.Background(), documentID)
	if err == nil {
		log.Printf("Cache hit for doc %s. Re-creating hub from Redis state.", documentID)
	} else if err == redis.Nil {
		log.Printf("Cache miss for doc %s. Loading from PostgreSQL.", documentID)
		doc, err := m.Store.GetDocument(documentID)
		if err != nil {
			return nil, err
		}
		content = doc.Content
		version = doc.Version

		if err := m.Cache.SetDocumentState(context.Background(), documentID, content, version); err != nil {
			log.Printf("WARN: Failed to prime cache for doc %s: %v", documentID, err)
		}
	} else {
		return nil, err
	}

	// =========================================================
	// === THIS IS THE LINE THAT WAS FIXED =====================
	// =========================================================
	hub := newHub(documentID, content, version, m) // Pass the manager itself, not its fields.

	m.hubs[documentID] = hub
	go hub.run()
	log.Printf("Hub created for document %s at version %d", documentID, version)
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

	hub, err := m.getOrCreateHub(documentID)
	if err != nil {
		http.Error(w, "Document not found or internal error", http.StatusNotFound)
		return
	}
	
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade connection:", err)
		return
	}

	client := &Client{
		ID:   uuid.NewString(),
		hub:  hub,
		conn: conn,
		send: make(chan *ServerMessage, 256),
	}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()

	log.Printf("Client %s (for user %s) connected to hub for document %s", client.ID, userID, documentID)
}