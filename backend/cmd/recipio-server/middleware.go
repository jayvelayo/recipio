package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/jayvelayo/recipio/internal/authn"
)

type contextKey string

const userIDKey contextKey = "userID"

type statusRecorder struct {
	http.ResponseWriter
	status int
	body   strings.Builder
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status >= 400 {
		r.body.Write(b)
	}
	return r.ResponseWriter.Write(b)
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		if rec.status >= 400 {
			log.Printf("ERROR %d %s %s: %s", rec.status, r.Method, r.URL.Path, strings.TrimSpace(rec.body.String()))
		}
	})
}

func withAuth(authDB authn.AuthDatabase, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow pre-injected userID (e.g. in tests)
		if _, ok := r.Context().Value(userIDKey).(string); ok {
			next.ServeHTTP(w, r)
			return
		}
		if authDB == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")
		userID, err := authDB.GetUserIDBySessionToken(token)
		if err != nil {
			http.Error(w, "Invalid or expired session", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func withCORS(allowedOrigins []string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Check if the origin is in the allowed list
		corsAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				corsAllowed = true
				break
			}
		}

		// Only log if CORS is denied
		if !corsAllowed {
			log.Printf("CORS denied: origin=%s, path=%s", origin, r.URL.Path)
		}

		if corsAllowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin") // Important: informs caches that response varies by Origin
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "86400") // Cache preflight for 24 hours
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		handler.ServeHTTP(w, r)
	})
}
