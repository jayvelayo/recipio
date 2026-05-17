package authn

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
)

type mockGoogleAuthDatabase struct {
	users    map[string]uuid.UUID
	googleIDs map[string]string
	byGoogle map[string]string
}

func newMockGoogleAuthDatabase() *mockGoogleAuthDatabase {
	return &mockGoogleAuthDatabase{
		users:    make(map[string]uuid.UUID),
		googleIDs: make(map[string]string),
		byGoogle: make(map[string]string),
	}
}

func (m *mockGoogleAuthDatabase) CreateUser(name, email string) (uuid.UUID, error) {
	id := uuid.New()
	m.users[email] = id
	return id, nil
}

func (m *mockGoogleAuthDatabase) GetUserByID(userID string) (User, error) {
	for email, id := range m.users {
		if id.String() == userID {
			return User{ID: id, Email: email}, nil
		}
	}
	return User{}, fmt.Errorf("user not found")
}

func (m *mockGoogleAuthDatabase) CreateSession(userID string) (string, error) {
	return "session-" + userID, nil
}

func (m *mockGoogleAuthDatabase) GetUserIDBySessionToken(token string) (string, error) {
	return "", nil
}

func (m *mockGoogleAuthDatabase) GetUserIDByEmail(email string) (uuid.UUID, error) {
	if id, ok := m.users[email]; ok {
		return id, nil
	}
	return uuid.Nil, nil
}

func (m *mockGoogleAuthDatabase) GetGoogleIDByUserID(userID string) (string, error) {
	return m.googleIDs[userID], nil
}

func (m *mockGoogleAuthDatabase) StoreGoogleID(userID, googleID string) error {
	m.googleIDs[userID] = googleID
	m.byGoogle[googleID] = userID
	return nil
}

func (m *mockGoogleAuthDatabase) GetUserIDByGoogleID(googleID string) (string, error) {
	return m.byGoogle[googleID], nil
}

func TestGoogleCreateCredentials(t *testing.T) {
	t.Run("Creates user and stores Google ID", func(t *testing.T) {
		db := newMockGoogleAuthDatabase()
		auth := GoogleAuthenticator{DB: db}
		err := auth.CreateCredentials("Alice", "alice@example.com", "google-id-123")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		userID, _ := db.GetUserIDByEmail("alice@example.com")
		if userID == uuid.Nil {
			t.Error("Expected user to be created")
		}
		storedGoogleID, _ := db.GetGoogleIDByUserID(userID.String())
		if storedGoogleID != "google-id-123" {
			t.Errorf("Expected Google ID to be stored, got %q", storedGoogleID)
		}
	})

	t.Run("Returns error for duplicate email", func(t *testing.T) {
		db := newMockGoogleAuthDatabase()
		auth := GoogleAuthenticator{DB: db}
		auth.CreateCredentials("Alice", "alice@example.com", "google-id-123")
		err := auth.CreateCredentials("Alice2", "alice@example.com", "google-id-456")
		if err == nil {
			t.Error("Expected error for duplicate email, got nil")
		}
	})
}

func TestVerifyGoogleID(t *testing.T) {
	db := newMockGoogleAuthDatabase()
	auth := GoogleAuthenticator{DB: db}
	auth.CreateCredentials("Bob", "bob@example.com", "bob-google-id")

	t.Run("Returns session token for known Google ID", func(t *testing.T) {
		token, err := auth.VerifyGoogleID("bob-google-id")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if token == "" {
			t.Error("Expected a session token, got empty string")
		}
	})

	t.Run("Returns error for unknown Google ID", func(t *testing.T) {
		_, err := auth.VerifyGoogleID("unknown-google-id")
		if err == nil {
			t.Error("Expected error for unknown Google ID, got nil")
		}
	})
}
