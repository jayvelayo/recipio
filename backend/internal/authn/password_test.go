package authn

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
)

type mockPasswordDatabase struct {
	users     map[string]uuid.UUID
	passwords map[string]string
}

func newMockPasswordDatabase() *mockPasswordDatabase {
	return &mockPasswordDatabase{
		users:     make(map[string]uuid.UUID),
		passwords: make(map[string]string),
	}
}

func (m *mockPasswordDatabase) CreateUser(name, email string) (uuid.UUID, error) {
	id := uuid.New()
	m.users[email] = id
	return id, nil
}

func (m *mockPasswordDatabase) CreateSession(userID string) (string, error) {
	return "session-" + userID, nil
}

func (m *mockPasswordDatabase) GetUserIDBySessionToken(token string) (string, error) {
	return "", nil
}

func (m *mockPasswordDatabase) GetUserIDByEmail(email string) (uuid.UUID, error) {
	if id, ok := m.users[email]; ok {
		return id, nil
	}
	return uuid.Nil, nil
}

func (m *mockPasswordDatabase) GetPasswordHash(userID string) (string, error) {
	if hash, ok := m.passwords[userID]; ok {
		return hash, nil
	}
	return "", fmt.Errorf("password not found")
}

func (m *mockPasswordDatabase) StorePasswordHash(userID, hash string) error {
	m.passwords[userID] = hash
	return nil
}

func TestPasswordCreateCredentials(t *testing.T) {
	t.Run("Creates user and stores password hash", func(t *testing.T) {
		db := newMockPasswordDatabase()
		auth := PasswordAuthenticator{DB: db}
		err := auth.CreateCredentials("Alice", "alice@example.com", "password123")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		userID, _ := db.GetUserIDByEmail("alice@example.com")
		if userID == uuid.Nil {
			t.Error("Expected user to be created")
		}
		if _, ok := db.passwords[userID.String()]; !ok {
			t.Error("Expected password hash to be stored")
		}
	})

	t.Run("Returns error for duplicate email", func(t *testing.T) {
		db := newMockPasswordDatabase()
		auth := PasswordAuthenticator{DB: db}
		auth.CreateCredentials("Alice", "alice@example.com", "password123")
		err := auth.CreateCredentials("Alice2", "alice@example.com", "password456")
		if err == nil {
			t.Error("Expected error for duplicate email, got nil")
		}
	})
}

func TestVerifyPassword(t *testing.T) {
	db := newMockPasswordDatabase()
	auth := PasswordAuthenticator{DB: db}
	auth.CreateCredentials("Bob", "bob@example.com", "correctpassword")

	t.Run("Returns session token for correct password", func(t *testing.T) {
		token, err := auth.VerifyPassword("bob@example.com", "correctpassword")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if token == "" {
			t.Error("Expected a session token, got empty string")
		}
	})

	t.Run("Returns error for incorrect password", func(t *testing.T) {
		_, err := auth.VerifyPassword("bob@example.com", "wrongpassword")
		if err == nil {
			t.Error("Expected error for wrong password, got nil")
		}
	})

	t.Run("Returns error for non-existent user", func(t *testing.T) {
		_, err := auth.VerifyPassword("unknown@example.com", "password")
		if err == nil {
			t.Error("Expected error for unknown user, got nil")
		}
	})
}
