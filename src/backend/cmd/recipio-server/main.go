package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	rec "github.com/jayvelayo/recipio/internal/recipes"
	"github.com/jayvelayo/recipio/internal/sqlite_db"
)

// loadEnvFile loads environment variables from .env file
func loadEnvFile() error {
	file, err := os.Open(".env")
	if err != nil {
		// .env file doesn't exist, that's okay
		return nil
	}
	defer file.Close()

	// Read file line by line
	content, err := os.ReadFile(".env")
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove quotes if present
			value = strings.Trim(value, `"'`)
			os.Setenv(key, value)
		}
	}

	return nil
}

// contains checks if a string slice contains a specific item
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// newServer creates and returns the HTTP server handler with all routes
func newServer(
	recipeDatabase rec.RecipeDatabase,
	allowedOrigins []string,
) http.Handler {
	mux := http.NewServeMux()
	SetUpRoutes(mux, recipeDatabase, allowedOrigins)
	return mux
}

// main entry point
func main() {
	// Load environment variables from .env file
	if err := loadEnvFile(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Read allowed origins from environment variable
	allowedOriginsStr := os.Getenv("ALLOWED_ORIGINS")
	var allowedOrigins []string

	// Always include localhost origins for development
	localhostOrigins := []string{
		"http://127.0.0.1:4002",
		"https://127.0.0.1:4002",
		"http://localhost:4002",
		"https://localhost:4002",
		"http://127.0.0.1:5173", // Vite dev server
		"https://127.0.0.1:5173",
		"http://localhost:5173",
		"https://localhost:5173",
	}

	if allowedOriginsStr == "" {
		// Default to localhost origins for development
		allowedOrigins = localhostOrigins
	} else {
		// Start with localhost origins, then add configured ones
		allowedOrigins = append(allowedOrigins, localhostOrigins...)

		// Parse comma-separated list of additional origins
		for _, origin := range strings.Split(allowedOriginsStr, ",") {
			trimmed := strings.TrimSpace(origin)
			if trimmed != "" && !contains(allowedOrigins, trimmed) {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
	}

	log.Printf("Allowed CORS origins: %v", allowedOrigins)

	// Initialize database
	recipeDb, err := sqlite_db.InitDb("recipes.db")
	if err != nil {
		log.Fatalf("unable to init db: %v", err)
	}
	defer recipeDb.CloseDb()

	// Create server and start listening
	srv := newServer(recipeDb, allowedOrigins)

	log.Println("Starting server on :4002...")
	if err := http.ListenAndServe(":4002", srv); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
