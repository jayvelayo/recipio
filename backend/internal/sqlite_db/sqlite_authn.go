package sqlite_db

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ===== Common Auth functions =====

func (db *SqliteDatabaseContext) CreateUser(name, email string) (uuid.UUID, error) {
	id := uuid.New()
	_, err := db.sqliteDb.Exec(
		"INSERT INTO users (id, name, email, created) VALUES (?, ?, ?, ?)",
		id.String(), name, email, time.Now(),
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create user: %w", err)
	}
	return id, nil
}

func (db *SqliteDatabaseContext) CreateSession(userID string) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("failed to generate session token: %w", err)
	}
	rawToken := hex.EncodeToString(raw)
	hash := sha256.Sum256([]byte(rawToken))
	hashedToken := hex.EncodeToString(hash[:])

	expiry := time.Now().Add(30 * 24 * time.Hour)
	_, err := db.sqliteDb.Exec(
		"INSERT INTO sessions (token, user_id, expires) VALUES (?, ?, ?)",
		hashedToken, userID, expiry,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	return rawToken, nil
}

func (db *SqliteDatabaseContext) GetUserIDBySessionToken(sessionToken string) (string, error) {
	hash := sha256.Sum256([]byte(sessionToken))
	hashedToken := hex.EncodeToString(hash[:])

	var userID string
	err := db.sqliteDb.QueryRow(
		"SELECT user_id FROM sessions WHERE token = ? AND expires > ?",
		hashedToken, time.Now(),
	).Scan(&userID)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("session not found or expired")
	}
	if err != nil {
		return "", err
	}
	return userID, nil
}

func (db *SqliteDatabaseContext) GetUserIDByEmail(email string) (uuid.UUID, error) {
	var idStr string
	err := db.sqliteDb.QueryRow(
		"SELECT id FROM users WHERE email = ?", email,
	).Scan(&idStr)
	if err == sql.ErrNoRows {
		return uuid.Nil, nil
	}
	if err != nil {
		return uuid.Nil, err
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid uuid in db: %w", err)
	}
	return id, nil
}

// ===== Password =====

func (db *SqliteDatabaseContext) GetPasswordHash(userID string) (string, error) {
	var hash string
	err := db.sqliteDb.QueryRow(
		"SELECT password FROM credentials WHERE user_id = ?", userID,
	).Scan(&hash)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("no credentials for user %s", userID)
	}
	if err != nil {
		return "", err
	}
	return hash, nil
}

func (db *SqliteDatabaseContext) StorePasswordHash(userID string, hash string) error {
	_, err := db.sqliteDb.Exec(
		"INSERT INTO credentials (user_id, password) VALUES (?, ?)", userID, hash,
	)
	if err != nil {
		return fmt.Errorf("failed to store password hash: %w", err)
	}
	return nil
}

// ===== Google =====

func (db *SqliteDatabaseContext) GetGoogleIDByUserID(userID string) (string, error) {
	var sub string
	err := db.sqliteDb.QueryRow(
		"SELECT sub FROM oauth WHERE user_id = ? AND provider = 'google'", userID,
	).Scan(&sub)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return sub, nil
}

func (db *SqliteDatabaseContext) StoreGoogleID(userID string, googleID string) error {
	_, err := db.sqliteDb.Exec(
		"INSERT INTO oauth (user_id, provider, sub) VALUES (?, 'google', ?)", userID, googleID,
	)
	if err != nil {
		return fmt.Errorf("failed to store google id: %w", err)
	}
	return nil
}

func (db *SqliteDatabaseContext) GetUserIDByGoogleID(googleID string) (string, error) {
	var userID string
	err := db.sqliteDb.QueryRow(
		"SELECT user_id FROM oauth WHERE provider = 'google' AND sub = ?", googleID,
	).Scan(&userID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return userID, nil
}
