package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
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

	allowedOrigins := []string{
		"http://127.0.0.1:4002",
		"https://127.0.0.1:4002",
		"http://localhost:4002",
		"https://localhost:4002",
	}

	// Initialize database
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatalf("unable to get user cache dir: %v", err)
	}
	dbPath := filepath.Join(cacheDir, "recipio", "recipes.db")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		log.Fatalf("unable to create cache dir: %v", err)
	}
	log.Printf("Using database: %s", dbPath)
	recipeDb, err := sqlite_db.InitDb(dbPath)
	if err != nil {
		log.Fatalf("unable to init db: %v", err)
	}
	defer recipeDb.CloseDb()

	port := os.Getenv("PORT")
	if port == "" {
		port = "4002"
	}

	srv := withLogging(newServer(recipeDb, allowedOrigins))
	log.Printf("Starting server on :%s...", port)
	if err := http.ListenAndServe(":"+port, srv); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
