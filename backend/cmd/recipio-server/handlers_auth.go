package main

import (
	"net/http"
	"strings"

	"github.com/jayvelayo/recipio/internal/authn"
)

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

func handlePasswordRegister(authDB authn.PasswordDatabase) http.Handler {
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
		w.WriteHeader(http.StatusCreated)
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
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}
		encodeJson(w, http.StatusOK, LoginResponse{Token: token, Email: body.Email})
	})
}
