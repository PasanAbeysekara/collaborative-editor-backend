package storage

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
}

type Document struct {
	ID      string
	Title   string
	OwnerID string
}

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

func (s *MemoryStore) CreateUser(email, password string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, u := range s.users {
		if u.Email == email {
			return nil, fmt.Errorf("user with email %s already exists", email)
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("could not hash password: %w", err)
	}

	user := User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: string(hashedPassword),
	}
	s.users[user.ID] = user
	return &user, nil
}

func (s *MemoryStore) GetUserByEmail(email string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.Email == email {
			return &user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (s *MemoryStore) GetUserByID(id string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return &user, nil
}

func (s *MemoryStore) CreateDocument(title, ownerID string) (*Document, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	doc := Document{
		ID:      uuid.NewString(),
		Title:   title,
		OwnerID: ownerID,
	}
	s.documents[doc.ID] = doc

	return &doc, nil
}
