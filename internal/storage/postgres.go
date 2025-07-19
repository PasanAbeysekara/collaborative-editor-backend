package storage

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) CreateUser(email, password string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{Email: email, PasswordHash: string(hashedPassword)}
	query := `INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`

	err = s.pool.QueryRow(context.Background(), query, email, string(hashedPassword)).Scan(&user.ID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *PostgresStore) GetUserByEmail(email string) (*User, error) {
	user := &User{}
	query := `SELECT id, email, password_hash FROM users WHERE email = $1`

	err := s.pool.QueryRow(context.Background(), query, email).Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *PostgresStore) CreateDocument(title, ownerID string) (*Document, error) {
	doc := &Document{Title: title, OwnerID: ownerID}
	query := `INSERT INTO documents (title, owner_id) VALUES ($1, $2) RETURNING id`

	err := s.pool.QueryRow(context.Background(), query, title, ownerID).Scan(&doc.ID)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func (s *PostgresStore) GetUserByID(id string) (*User, error) {
	user := &User{}
	query := `SELECT id, email, password_hash FROM users WHERE id = $1`

	err := s.pool.QueryRow(context.Background(), query, id).Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *PostgresStore) CheckDocumentPermission(documentID, userID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM documents WHERE id = $1 AND owner_id = $2)`

	err := s.pool.QueryRow(context.Background(), query, documentID, userID).Scan(&exists)
	if err != nil {

		if err == pgx.ErrNoRows {
			return false, nil
		}

		return false, err
	}

	return exists, nil
}
