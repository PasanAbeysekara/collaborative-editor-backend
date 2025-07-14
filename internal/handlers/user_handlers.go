package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/your-username/collaborative-editor/internal/auth"
	"github.com/your-username/collaborative-editor/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	Store *storage.MemoryStore
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

// Register creates a new user.
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.Store.CreateUser(req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict) // 409 Conflict
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user) // Don't send password hash back in a real app!
}

// Login authenticates a user and returns a JWT.
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.Store.GetUserByEmail(req.Email)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Create JWT token
	token, err := auth.CreateJWT(user.ID, 24*time.Hour) // Token valid for 24 hours
	if err != nil {
		http.Error(w, "Could not create token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{Token: token})
}