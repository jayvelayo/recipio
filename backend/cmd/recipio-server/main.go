package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jayvelayo/recipio/internal/authn"
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
	authDatabase authn.PasswordDatabase,
	googleDB authn.GoogleAuthDatabase,
	googleCfg authn.GoogleOAuthConfig,
	allowedOrigins []string,
	emailSender authn.EmailSender,
) http.Handler {
	mux := http.NewServeMux()
	SetUpRoutes(mux, recipeDatabase, authDatabase, allowedOrigins, googleDB, googleCfg, emailSender)
	return mux
}

// main entry point
func main() {
	// Load environment variables from .env file
	if err := loadEnvFile(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "http://localhost:4002"
	}

	allowedOrigins := []string{
		"http://127.0.0.1:4002",
		"https://127.0.0.1:4002",
		"http://localhost:4002",
		"https://localhost:4002",
	}
	if appURL != "http://localhost:4002" && appURL != "https://localhost:4002" {
		allowedOrigins = append(allowedOrigins, appURL)
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
	// TODO: use Redis/PostgreDB for auth database as next learning step
	authDb, ok := recipeDb.(authn.PasswordDatabase)
	if !ok {
		log.Fatal("database does not implement authn.PasswordDatabase")
	}
	googleDb, _ := recipeDb.(authn.GoogleAuthDatabase)

	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if googleClientID == "" || googleClientSecret == "" {
		log.Fatal("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must be set")
	}
	googleRedirectURI := os.Getenv("GOOGLE_REDIRECT_URI")
	if googleRedirectURI == "" {
		googleRedirectURI = "http://localhost:4002/auth/google/callback"
	}
	googleCfg := authn.GoogleOAuthConfig{
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		RedirectURI:  googleRedirectURI,
	}

	resendAPIKey := os.Getenv("RESEND_API_KEY")
	if resendAPIKey == "" {
		log.Printf("Warning: RESEND_API_KEY not set, email verification disabled (accounts auto-verified)")
	}
	emailFrom := os.Getenv("EMAIL_FROM")
	if emailFrom == "" {
		emailFrom = "Recipio <onboarding@resend.dev>"
	}
	emailSender := authn.EmailSender{
		APIKey: resendAPIKey,
		From:   emailFrom,
		AppURL: appURL,
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "4002"
	}

	srv := withLogging(newServer(recipeDb, authDb, googleDb, googleCfg, allowedOrigins, emailSender))
	log.Printf("Starting server on :%s...", port)
	if err := http.ListenAndServe(":"+port, srv); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
