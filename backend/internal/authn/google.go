package authn

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

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
	user, err := a.DB.GetUserIDByEmail(email)
	if err != nil {
		return fmt.Errorf("Error occurred while fetching user %s: %w", email, err)
	}
	if user != uuid.Nil {
		return fmt.Errorf("User with this email already exists")
	}
	uuid, err := a.DB.CreateUser(name, email)
	if err != nil {
		return err
	}
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

// FindOrCreateSession returns a session token for the given Google user,
// creating a new account if one doesn't already exist for this Google ID.
func (a GoogleAuthenticator) FindOrCreateSession(info GoogleUserInfo) (string, error) {
	token, err := a.VerifyGoogleID(info.ID)
	if err == nil {
		return token, nil
	}
	if err := a.CreateCredentials(info.Name, info.Email, info.ID); err != nil {
		return "", err
	}
	return a.VerifyGoogleID(info.ID)
}

// GoogleOAuthConfig holds the OAuth 2.0 credentials and redirect URI for Google sign-in.
type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// GoogleUserInfo is the subset of fields returned by Google's userinfo endpoint.
type GoogleUserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type googleTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// GetAuthURL returns the Google OAuth authorization URL the user should be redirected to.
func (c GoogleOAuthConfig) GetAuthURL(state string) string {
	params := url.Values{
		"client_id":     {c.ClientID},
		"redirect_uri":  {c.RedirectURI},
		"response_type": {"code"},
		"scope":         {"openid email profile"},
		"state":         {state},
		"access_type":   {"online"},
	}
	return "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()
}

// ExchangeCodeForUserInfo trades an authorization code for an access token,
// then fetches the user's profile from Google's userinfo endpoint.
func (c GoogleOAuthConfig) ExchangeCodeForUserInfo(ctx context.Context, code string) (GoogleUserInfo, error) {
	params := url.Values{
		"code":          {code},
		"client_id":     {c.ClientID},
		"client_secret": {c.ClientSecret},
		"redirect_uri":  {c.RedirectURI},
		"grant_type":    {"authorization_code"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(params.Encode()))
	if err != nil {
		return GoogleUserInfo{}, fmt.Errorf("building token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return GoogleUserInfo{}, fmt.Errorf("token exchange: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return GoogleUserInfo{}, fmt.Errorf("token exchange status %d: %s", resp.StatusCode, body)
	}

	var tokenResp googleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return GoogleUserInfo{}, fmt.Errorf("decoding token response: %w", err)
	}

	userReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return GoogleUserInfo{}, fmt.Errorf("building userinfo request: %w", err)
	}
	userReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

	userResp, err := http.DefaultClient.Do(userReq)
	if err != nil {
		return GoogleUserInfo{}, fmt.Errorf("userinfo request: %w", err)
	}
	defer userResp.Body.Close()

	if userResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(userResp.Body)
		return GoogleUserInfo{}, fmt.Errorf("userinfo status %d: %s", userResp.StatusCode, body)
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(userResp.Body).Decode(&userInfo); err != nil {
		return GoogleUserInfo{}, fmt.Errorf("decoding userinfo: %w", err)
	}
	return userInfo, nil
}
