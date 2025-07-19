package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/pasanAbeysekara/collaborative-editor/internal/auth"
	"github.com/pasanAbeysekara/collaborative-editor/internal/storage"
)

type DocumentHandler struct {
	Store storage.Store
}

type CreateDocumentRequest struct {
	Title string `json:"title"`
}

func (h *DocumentHandler) CreateDocument(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		http.Error(w, "Could not get user ID from context", http.StatusInternalServerError)
		return
	}

	var req CreateDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		req.Title = "Untitled Document"
	}

	doc, err := h.Store.CreateDocument(req.Title, userID)
	if err != nil {
		http.Error(w, "Failed to create document", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(doc)
}
