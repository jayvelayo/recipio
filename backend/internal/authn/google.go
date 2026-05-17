package authn

import (
	"fmt"

	"github.com/google/uuid"
)

type GoogleAuthDatabase interface {
	AuthDatabase
	GetGoogleIDByUserID(userID string) (string, error)
	StoreGoogleID(userID string, googleID string) error
	GetUserIDByGoogleID(googleID string) (string, error)
}

type GoogleAuthenticator struct {
	DB GoogleAuthDatabase
}

func (a GoogleAuthenticator) CreateCredentials(name string, email string, googleID string) error {
	// Make sure the email hasn't been taken already
	user, err := a.DB.GetUserIDByEmail(email)
	if err != nil {
		return fmt.Errorf("Error occurred while fetching user %s: %w", email, err)
	}
	if user != uuid.Nil {
		return fmt.Errorf("User with this email already exists")
	}
	// Create the user
	uuid, err := a.DB.CreateUser(name, email)
	if err != nil {
		return err
	}
	// Store the google ID
	err = a.DB.StoreGoogleID(uuid.String(), googleID)
	if err != nil {
		return err
	}
	return nil
}

func (a GoogleAuthenticator) VerifyGoogleID(googleID string) (string, error) {
	userID, err := a.DB.GetUserIDByGoogleID(googleID)
	if err != nil {
		return "", err
	}
	if userID == "" {
		return "", fmt.Errorf("No user associated with this Google ID")
	}
	sessionToken, err := a.DB.CreateSession(userID)
	if err != nil {
		return "", fmt.Errorf("Error creating session: %w", err)
	}
	return sessionToken, nil
}
