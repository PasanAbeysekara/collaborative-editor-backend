package storage

type Store interface {
	CreateUser(email, password string) (*User, error)
	GetUserByEmail(email string) (*User, error)
	GetUserByID(id string) (*User, error)

	CheckDocumentPermission(documentID, userID string) (bool, error)
	CreateDocument(title, ownerID string) (*Document, error)
	GetDocument(documentID string) (*Document, error)
	GetUserDocuments(userID string) ([]*Document, error)
	UpdateDocument(documentID, content string, version int) error
	ShareDocument(documentID, ownerID, targetUserID, role string) error
}
