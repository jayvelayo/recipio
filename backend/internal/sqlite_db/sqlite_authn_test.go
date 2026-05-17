package sqlite_db_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jayvelayo/recipio/internal/authn"
	"github.com/jayvelayo/recipio/internal/sqlite_db"
)

func initAuthDB(t *testing.T) authn.PasswordDatabase {
	t.Helper()
	db, err := sqlite_db.InitDb(":memory:")
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	t.Cleanup(db.CloseDb)
	pd, ok := db.(authn.PasswordDatabase)
	if !ok {
		t.Fatal("db does not implement authn.PasswordDatabase")
	}
	return pd
}

func initGoogleAuthDB(t *testing.T) authn.GoogleAuthDatabase {
	t.Helper()
	db, err := sqlite_db.InitDb(":memory:")
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	t.Cleanup(db.CloseDb)
	gd, ok := db.(authn.GoogleAuthDatabase)
	if !ok {
		t.Fatal("db does not implement authn.GoogleAuthDatabase")
	}
	return gd
}

func TestCreateUser(t *testing.T) {
	db := initAuthDB(t)

	t.Run("creates a new user and returns a non-nil UUID", func(t *testing.T) {
		id, err := db.CreateUser("Alice", "alice@example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id == uuid.Nil {
			t.Error("expected non-nil UUID")
		}
	})

	t.Run("duplicate email returns error", func(t *testing.T) {
		db.CreateUser("Bob", "bob@example.com")
		_, err := db.CreateUser("Bob2", "bob@example.com")
		if err == nil {
			t.Error("expected error for duplicate email, got nil")
		}
	})
}

func TestGetUserIDByEmail(t *testing.T) {
	db := initAuthDB(t)

	t.Run("returns Nil UUID for unknown email", func(t *testing.T) {
		id, err := db.GetUserIDByEmail("nobody@example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != uuid.Nil {
			t.Errorf("expected uuid.Nil, got %v", id)
		}
	})

	t.Run("returns correct UUID for known email", func(t *testing.T) {
		created, _ := db.CreateUser("Carol", "carol@example.com")
		found, err := db.GetUserIDByEmail("carol@example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if found != created {
			t.Errorf("expected %v, got %v", created, found)
		}
	})
}

func TestCreateAndValidateSession(t *testing.T) {
	db := initAuthDB(t)
	userID, _ := db.CreateUser("Dave", "dave@example.com")

	t.Run("valid token resolves to user", func(t *testing.T) {
		token, err := db.CreateSession(userID.String())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if token == "" {
			t.Fatal("expected non-empty token")
		}
		gotUserID, err := db.GetUserIDBySessionToken(token)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotUserID != userID.String() {
			t.Errorf("expected %v, got %v", userID.String(), gotUserID)
		}
	})

	t.Run("invalid token returns error", func(t *testing.T) {
		_, err := db.GetUserIDBySessionToken("notavalidtoken")
		if err == nil {
			t.Error("expected error for invalid token, got nil")
		}
	})

	t.Run("each session produces a unique token", func(t *testing.T) {
		t1, _ := db.CreateSession(userID.String())
		t2, _ := db.CreateSession(userID.String())
		if t1 == t2 {
			t.Error("expected unique tokens per session")
		}
	})
}

func TestPasswordHash(t *testing.T) {
	db := initAuthDB(t)
	userID, _ := db.CreateUser("Eve", "eve@example.com")

	t.Run("stores and retrieves hash", func(t *testing.T) {
		if err := db.StorePasswordHash(userID.String(), "hashed_pw"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		hash, err := db.GetPasswordHash(userID.String())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if hash != "hashed_pw" {
			t.Errorf("expected 'hashed_pw', got %q", hash)
		}
	})

	t.Run("unknown user returns error", func(t *testing.T) {
		_, err := db.GetPasswordHash("nonexistent-id")
		if err == nil {
			t.Error("expected error for unknown user, got nil")
		}
	})
}

func TestGoogleOAuth(t *testing.T) {
	db := initGoogleAuthDB(t)
	userID, _ := db.CreateUser("Frank", "frank@example.com")

	if err := db.StoreGoogleID(userID.String(), "google-sub-123"); err != nil {
		t.Fatalf("unexpected error storing google id: %v", err)
	}

	t.Run("resolves user by google ID", func(t *testing.T) {
		gotUserID, err := db.GetUserIDByGoogleID("google-sub-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotUserID != userID.String() {
			t.Errorf("expected %v, got %v", userID.String(), gotUserID)
		}
	})

	t.Run("resolves google ID by user", func(t *testing.T) {
		sub, err := db.GetGoogleIDByUserID(userID.String())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sub != "google-sub-123" {
			t.Errorf("expected 'google-sub-123', got %q", sub)
		}
	})

	t.Run("unknown google ID returns empty string", func(t *testing.T) {
		gotUserID, err := db.GetUserIDByGoogleID("unknown-sub")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotUserID != "" {
			t.Errorf("expected empty string, got %q", gotUserID)
		}
	})
}
