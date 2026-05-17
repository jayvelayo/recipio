package authn

import (
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type PasswordDatabase interface {
	AuthDatabase
	GetPasswordHash(userID string) (string, error)
	StorePasswordHash(userID string, hash string) error
}

type PasswordAuthenticator struct {
	DB PasswordDatabase
}

var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy"), bcrypt.DefaultCost)

func (a PasswordAuthenticator) CreateCredentials(name string, email string, password string) error {
	// Make sure the email hasn't been taken already
	user, err := a.DB.GetUserIDByEmail(email)
	if err != nil {
		return fmt.Errorf("Error occurred while fetching user %s: %w", email, err)
	}
	if user != uuid.Nil {
		return fmt.Errorf("User with this email already exists")
	}
	uuid, err := a.DB.CreateUser(name, email)
	if err != nil {
		return fmt.Errorf("Error occurred while creating user %s: %w", email, err)
	}
	// Hash the password and store it
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("Error occurred while hashing password for user %s: %w", email, err)
	}
	err = a.DB.StorePasswordHash(uuid.String(), string(hashed))
	if err != nil {
		return fmt.Errorf("Error occurred while storing password for user %s: %w", email, err)
	}
	return nil
}

func (a PasswordAuthenticator) ChangePassword(userID string, newPassword string) error {
	return fmt.Errorf("Not implemented")
}

func (a PasswordAuthenticator) VerifyPassword(email string, password string) (string, error) {
	userID, err := a.DB.GetUserIDByEmail(email)
	if err != nil {
		return "", fmt.Errorf("Error occurred while fetching user %s: %w", email, err)
	}
	if userID == uuid.Nil {
		// Fake the time taken to compare the password hash to prevent user enumeration
		bcrypt.CompareHashAndPassword([]byte(dummyHash), []byte(password))
		return "", fmt.Errorf("User with this email does not exist")
	}
	hash, err := a.DB.GetPasswordHash(userID.String())
	if err != nil {
		return "", fmt.Errorf("Error occurred while fetching password hash for user %s: %w", email, err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return "", fmt.Errorf("Incorrect password")
	}
	sessionToken, err := a.DB.CreateSession(userID.String())
	if err != nil {
		return "", fmt.Errorf("Error occurred while creating session for user %s: %w", email, err)
	}
	return sessionToken, nil
}
