package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/jayvelayo/recipio/internal/authn"
	"golang.org/x/time/rate"
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

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type rateLimiterStore struct {
	mu       sync.Mutex
	limiters map[string]*ipLimiter
	lim      rate.Limit
	burst    int
}

func newRateLimiterStore(limit rate.Limit, burst int) *rateLimiterStore {
	s := &rateLimiterStore{
		limiters: make(map[string]*ipLimiter),
		lim:      limit,
		burst:    burst,
	}
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			s.mu.Lock()
			for ip, l := range s.limiters {
				if time.Since(l.lastSeen) > 30*time.Minute {
					delete(s.limiters, ip)
				}
			}
			s.mu.Unlock()
		}
	}()
	return s
}

func (s *rateLimiterStore) getLimiter(ip string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()
	if l, ok := s.limiters[ip]; ok {
		l.lastSeen = time.Now()
		return l.limiter
	}
	lim := rate.NewLimiter(s.lim, s.burst)
	s.limiters[ip] = &ipLimiter{limiter: lim, lastSeen: time.Now()}
	return lim
}

func withRateLimit(store *rateLimiterStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			if !store.getLimiter(ip).Allow() {
				http.Error(w, "Too many requests, please try again later", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
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
