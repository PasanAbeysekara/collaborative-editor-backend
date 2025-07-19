package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx"
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
	query := `
        SELECT EXISTS (
            SELECT 1 FROM documents WHERE id = $1 AND owner_id = $2
            UNION ALL
            SELECT 1 FROM document_permissions WHERE document_id = $1 AND user_id = $2
        )
    `

	err := s.pool.QueryRow(context.Background(), query, documentID, userID).Scan(&exists)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return exists, nil
}

func (s *PostgresStore) GetDocument(documentID string) (*Document, error) {
	doc := &Document{}
	query := `SELECT id, title, owner_id, content, version FROM documents WHERE id = $1`

	err := s.pool.QueryRow(context.Background(), query, documentID).Scan(
		&doc.ID, &doc.Title, &doc.OwnerID, &doc.Content, &doc.Version,
	)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func (s *PostgresStore) UpdateDocument(documentID, content string, version int) error {
	query := `UPDATE documents SET content = $1, version = $2 WHERE id = $3`

	result, err := s.pool.Exec(context.Background(), query, content, version, documentID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (s *PostgresStore) ShareDocument(documentID, ownerID, targetUserID, role string) error {

	// Use a transaction to ensure atomicity:
	// 1. Verify the person sharing is the owner.
	// 2. Insert the permission.
	tx, err := s.pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background()) // Rollback on error

	// 1. Verify ownership.
	var isOwner bool
	ownerCheckQuery := `SELECT EXISTS(SELECT 1 FROM documents WHERE id = $1 AND owner_id = $2)`
	err = tx.QueryRow(context.Background(), ownerCheckQuery, documentID, ownerID).Scan(&isOwner)
	if err != nil {
		return err
	}
	if !isOwner {
		return fmt.Errorf("permission denied: only the owner can share this document")
	}

	// 2. Insert the permission, ignoring conflicts if it already exists.
	// "ON CONFLICT (document_id, user_id) DO NOTHING" is an "upsert" that prevents errors
	// if you try to share with the same person twice.
	insertQuery := `
        INSERT INTO document_permissions (document_id, user_id, role)
        VALUES ($1, $2, $3)
        ON CONFLICT (document_id, user_id) DO NOTHING
    `
	_, err = tx.Exec(context.Background(), insertQuery, documentID, targetUserID, role)
	if err != nil {
		return err
	}

	return tx.Commit(context.Background()) // Commit the transaction
}
