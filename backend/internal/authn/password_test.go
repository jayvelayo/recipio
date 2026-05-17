package authn

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

type mockPasswordDatabase struct {
	users      map[string]uuid.UUID
	userNames  map[string]string
	userEmails map[string]string
	passwords  map[string]string
}

func newMockPasswordDatabase() *mockPasswordDatabase {
	return &mockPasswordDatabase{
		users:      make(map[string]uuid.UUID),
		userNames:  make(map[string]string),
		userEmails: make(map[string]string),
		passwords:  make(map[string]string),
	}
}

func (m *mockPasswordDatabase) CreateUser(name, email string) (uuid.UUID, error) {
	id := uuid.New()
	m.users[email] = id
	m.userNames[id.String()] = name
	m.userEmails[id.String()] = email
	return id, nil
}

func (m *mockPasswordDatabase) GetUserByID(userID string) (User, error) {
	name, ok := m.userNames[userID]
	if !ok {
		return User{}, fmt.Errorf("user not found")
	}
	id, _ := uuid.Parse(userID)
	return User{ID: id, Name: name, Email: m.userEmails[userID]}, nil
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

func (m *mockPasswordDatabase) CreateEmailVerification(userID, hashedToken string, expires time.Time) error {
	return nil
}

func (m *mockPasswordDatabase) GetUserIDByVerificationToken(hashedToken string) (string, error) {
	return "", nil
}

func (m *mockPasswordDatabase) MarkEmailVerified(userID string) error {
	return nil
}

func (m *mockPasswordDatabase) IsEmailVerified(userID string) (bool, error) {
	return true, nil
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
