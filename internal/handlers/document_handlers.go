package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/pasanAbeysekara/collaborative-editor/internal/auth"
	"github.com/pasanAbeysekara/collaborative-editor/internal/storage"
)

type DocumentHandler struct {
	Store storage.Store
}

type CreateDocumentRequest struct {
	Title string `json:"title"`
}

type ShareDocumentRequest struct {
	TargetUserEmail string `json:"email"`
	Role            string `json:"role"`
}

type UpdateDocumentRequest struct {
	Content string `json:"content"`
	Version int    `json:"version"`
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

func (h *DocumentHandler) ShareDocument(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		http.Error(w, "Could not get user ID from context", http.StatusInternalServerError)
		return
	}

	documentID := chi.URLParam(r, "documentID")

	var req ShareDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	targetUser, err := h.Store.GetUserByEmail(req.TargetUserEmail)
	if err != nil {
		http.Error(w, "Target user not found", http.StatusNotFound)
		return
	}

	err = h.Store.ShareDocument(documentID, ownerID, targetUser.ID, req.Role)
	if err != nil {
		if err.Error() == "permission denied: only the owner can share this document" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, "Failed to share document", http.StatusInternalServerError)
		log.Printf("Error sharing document: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Document %s shared with %s successfully", documentID, req.TargetUserEmail)
}

func (h *DocumentHandler) CheckPermission(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	documentID := chi.URLParam(r, "documentID")

	hasPermission, err := h.Store.CheckDocumentPermission(documentID, userID)
	if err != nil {
		http.Error(w, "Internal check failed", http.StatusInternalServerError)
		return
	}

	if !hasPermission {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *DocumentHandler) GetDocument(w http.ResponseWriter, r *http.Request) {
    documentID := chi.URLParam(r, "documentID")

    doc, err := h.Store.GetDocument(documentID)
    if err != nil {
        if err == pgx.ErrNoRows {
            http.Error(w, "Document not found", http.StatusNotFound)
            return
        }
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

	// Return the document details as JSON.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
}

func (h *DocumentHandler) SaveDocument(w http.ResponseWriter, r *http.Request) {
      documentID := chi.URLParam(r, "documentID")
      
      var req UpdateDocumentRequest
      if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
          http.Error(w, "Invalid request body", http.StatusBadRequest)
          return
      }

      err := h.Store.UpdateDocument(documentID, req.Content, req.Version)
      if err != nil {
          http.Error(w, "Failed to update document", http.StatusInternalServerError)
          return
      }
      w.WriteHeader(http.StatusOK)
  }