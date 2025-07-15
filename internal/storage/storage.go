package storage

type Store interface {
	CreateUser(email, password string) (*User, error)
	GetUserByEmail(email string) (*User, error)
	CreateDocument(title, ownerID string) (*Document, error)
	GetUserByID(id string) (*User, error)
}
