package storage

import (
	"fmt"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user in our system
type User struct {
	ID           string
	Email        string
	PasswordHash string // Never store plain text passwords!
}

// Document represents document metadata
type Document struct {
	ID      string
	Title   string
	OwnerID string
}

// MemoryStore holds data in memory. Thread-safe.
type MemoryStore struct {
	mu        sync.RWMutex
	users     map[string]User
	documents map[string]Document
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users:     make(map[string]User),
		documents: make(map[string]Document),
	}
}

// -- User Methods --

func (s *MemoryStore) CreateUser(email, password string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[email]; exists {
		return nil, fmt.Errorf("user with email %s already exists", email)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("could not hash password: %w", err)
	}

	// In a real app, ID would be a UUID.
	id := fmt.Sprintf("user_%d", len(s.users)+1)
	user := User{
		ID:           id,
		Email:        email,
		PasswordHash: string(hashedPassword),
	}
	s.users[email] = user
	return &user, nil
}

func (s *MemoryStore) GetUserByEmail(email string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[email]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return &user, nil
}