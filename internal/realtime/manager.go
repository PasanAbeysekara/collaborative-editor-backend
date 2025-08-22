package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		
		if origin == "" {
			return true
		}
		
		allowedOrigins := []string{
			"https://solid-guide-974w6599qrjgfpr9j-5174.app.github.dev",
			"http://localhost:3000",
			"http://localhost:3001",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:3001",
		}

		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				return true
			}
			// Support wildcard subdomains like *.trycloudflare.com
			if strings.HasSuffix(allowedOrigin, ".trycloudflare.com") && strings.HasSuffix(origin, ".trycloudflare.com") {
				return true
			}
			if strings.HasSuffix(allowedOrigin, ".ngrok.io") && strings.HasSuffix(origin, ".ngrok.io") {
				return true
			}
		}

		// Allow development environments
		return strings.HasPrefix(origin, "http://localhost:") ||
			strings.HasPrefix(origin, "http://127.0.0.1:")
	},
}

type Manager struct {
	hubs               map[string]*Hub
	mu                 sync.RWMutex
	Cache              storage.Cache
	documentServiceURL string
}

func NewManager(cache storage.Cache, docServiceURL string) *Manager {
	return &Manager{
		hubs:               make(map[string]*Hub),
		Cache:              cache,
		documentServiceURL: docServiceURL,
	}
}

func (m *Manager) getDocumentFromService(documentID string) (*storage.Document, error) {
	url := fmt.Sprintf("%s/documents/%s", m.documentServiceURL, documentID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call document service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("document service returned status %d", resp.StatusCode)
	}

	var doc storage.Document
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, fmt.Errorf("failed to decode document from service: %w", err)
	}

	return &doc, nil
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
		doc, err := m.getDocumentFromService(documentID)
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

	hub := newHub(documentID, content, version, m)

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

	hasPermission, err := m.checkPermissions(documentID, userID)
	if err != nil {
		log.Printf("Error calling document-service for permissions: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if !hasPermission {
		http.Error(w, "Forbidden", http.StatusForbidden)
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

func (m *Manager) checkPermissions(documentID, userID string) (bool, error) {
	url := fmt.Sprintf("%s/documents/%s/permissions/%s", m.documentServiceURL, documentID, userID)
	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}
