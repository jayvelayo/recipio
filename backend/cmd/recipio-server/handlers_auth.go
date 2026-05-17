package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jayvelayo/recipio/internal/authn"
)

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func handleGoogleLogin(cfg authn.GoogleOAuthConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state, err := generateState()
		if err != nil {
			http.Error(w, "Failed to initiate login", http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "google_oauth_state",
			Value:    state,
			MaxAge:   300,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		http.Redirect(w, r, cfg.GetAuthURL(state), http.StatusTemporaryRedirect)
	})
}

func handleGoogleCallback(cfg authn.GoogleOAuthConfig, db authn.GoogleAuthDatabase) http.Handler {
	auth := authn.GoogleAuthenticator{DB: db}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stateCookie, err := r.Cookie("google_oauth_state")
		if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
			http.Error(w, "Invalid state parameter", http.StatusBadRequest)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:   "google_oauth_state",
			MaxAge: -1,
		})

		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Missing authorization code", http.StatusBadRequest)
			return
		}

		userInfo, err := cfg.ExchangeCodeForUserInfo(r.Context(), code)
		if err != nil {
			http.Error(w, "Failed to authenticate with Google", http.StatusInternalServerError)
			return
		}

		token, err := auth.FindOrCreateSession(userInfo)
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				http.Error(w, "An account with this email already exists. Please log in with your password.", http.StatusConflict)
			} else {
				http.Error(w, "Failed to create session", http.StatusInternalServerError)
			}
			return
		}

		http.Redirect(w, r, "/#google_token="+token, http.StatusTemporaryRedirect)
	})
}

func handleGetUserInfo(authDB authn.PasswordDatabase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")
		userID, err := authDB.GetUserIDBySessionToken(token)
		if err != nil {
			http.Error(w, "Invalid or expired session", http.StatusUnauthorized)
			return
		}
		user, err := authDB.GetUserByID(userID)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		encodeJson(w, http.StatusOK, UserInfoResponse{Name: user.Name, Email: user.Email})
	})
}

func handlePasswordRegister(authDB authn.PasswordDatabase, sender authn.EmailSender) http.Handler {
	auth := authn.PasswordAuthenticator{DB: authDB}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}
		body, err := decodeJson[RegisterRequest](r)
		if err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}
		if body.Name == "" || body.Email == "" || body.Password == "" {
			http.Error(w, "name, email, and password are required", http.StatusBadRequest)
			return
		}
		if err := auth.CreateCredentials(body.Name, body.Email, body.Password); err != nil {
			if strings.Contains(err.Error(), "already exists") {
				http.Error(w, err.Error(), http.StatusConflict)
			} else {
				http.Error(w, "Failed to create account", http.StatusInternalServerError)
			}
			return
		}

		if sender.APIKey == "" {
			// No email service configured: auto-verify for local dev.
			userID, err := authDB.GetUserIDByEmail(body.Email)
			if err == nil {
				authDB.MarkEmailVerified(userID.String())
			}
			w.WriteHeader(http.StatusCreated)
			return
		}

		userID, err := authDB.GetUserIDByEmail(body.Email)
		if err != nil {
			http.Error(w, "Failed to create account", http.StatusInternalServerError)
			return
		}
		rawBytes := make([]byte, 32)
		rand.Read(rawBytes)
		rawToken := hex.EncodeToString(rawBytes)
		hash := sha256.Sum256([]byte(rawToken))
		hashedToken := hex.EncodeToString(hash[:])
		expires := time.Now().Add(24 * time.Hour)
		if err := authDB.CreateEmailVerification(userID.String(), hashedToken, expires); err != nil {
			http.Error(w, "Failed to create account", http.StatusInternalServerError)
			return
		}
		if err := sender.SendVerificationEmail(body.Email, body.Name, rawToken); err != nil {
			log.Printf("ERROR sending verification email to %s: %v", body.Email, err)
			http.Error(w, "Failed to send verification email", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	})
}

func handleEmailVerification(authDB authn.PasswordDatabase, appURL string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawToken := r.URL.Query().Get("token")
		if rawToken == "" {
			http.Error(w, "Missing token", http.StatusBadRequest)
			return
		}
		hash := sha256.Sum256([]byte(rawToken))
		hashedToken := hex.EncodeToString(hash[:])
		userID, err := authDB.GetUserIDByVerificationToken(hashedToken)
		if err != nil {
			http.Error(w, "Invalid or expired verification link", http.StatusBadRequest)
			return
		}
		if err := authDB.MarkEmailVerified(userID); err != nil {
			http.Error(w, "Failed to verify email", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, appURL+"/?email_verified=true", http.StatusTemporaryRedirect)
	})
}

func handlePasswordLogin(authDB authn.PasswordDatabase) http.Handler {
	auth := authn.PasswordAuthenticator{DB: authDB}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}
		body, err := decodeJson[LoginRequest](r)
		if err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}
		if body.Email == "" || body.Password == "" {
			http.Error(w, "email and password are required", http.StatusBadRequest)
			return
		}
		token, err := auth.VerifyPassword(body.Email, body.Password)
		if err != nil {
			if errors.Is(err, authn.ErrEmailNotVerified) {
				http.Error(w, "Please verify your email address before signing in", http.StatusForbidden)
				return
			}
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}
		encodeJson(w, http.StatusOK, LoginResponse{Token: token, Email: body.Email})
	})
}
