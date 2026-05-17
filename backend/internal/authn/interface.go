package authn

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID      uuid.UUID
	Name    string
	Email   string
	Created time.Time
}

type Authenticator interface {
	Authenticate(token string) (userID string, err error)
}

type AuthDatabase interface {
	CreateUser(name string, email string) (uuid.UUID, error)
	CreateSession(userID string) (sessionToken string, err error)
	GetUserIDBySessionToken(sessionToken string) (userID string, err error)
	GetUserIDByEmail(email string) (uuid.UUID, error)
	GetUserByID(userID string) (User, error)
}
