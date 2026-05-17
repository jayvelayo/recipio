package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/jayvelayo/recipio/internal/authn"
	rec "github.com/jayvelayo/recipio/internal/recipes"
)

// SetUpRoutes registers all API endpoints and static file serving
// This is the single place to see the entire API contract
func SetUpRoutes(
	mux *http.ServeMux,
	recipeDatabase rec.RecipeDatabase,
	authDatabase authn.PasswordDatabase,
	allowedOrigins []string,
	googleDB authn.GoogleAuthDatabase,
	googleCfg authn.GoogleOAuthConfig,
) {
	setupSPAHandler(mux)

	protected := func(h http.Handler) http.Handler {
		return withCORS(allowedOrigins, withAuth(authDatabase, h))
	}
	cors := func(h http.Handler) http.Handler {
		return withCORS(allowedOrigins, h)
	}
	preflight := cors(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	// Recipe endpoints
	mux.Handle("POST /recipes", protected(handleDesignCreateRecipe(recipeDatabase)))
	mux.Handle("GET /recipes/{id}", protected(handleDesignGetRecipe(recipeDatabase)))
	mux.Handle("GET /recipes", protected(handleDesignGetAllRecipes(recipeDatabase)))
	mux.Handle("PUT /recipes/{id}", protected(handleDesignUpdateRecipe(recipeDatabase)))
	mux.Handle("DELETE /recipes/{id}", protected(handleDesignDeleteRecipe(recipeDatabase)))
	mux.Handle("OPTIONS /recipes", preflight)
	mux.Handle("OPTIONS /recipes/{id}", preflight)

	// AI parsing endpoint
	mux.Handle("POST /parse-recipe", cors(handleDesignParseRecipe(rec.NewAIParser(os.Getenv("GROQ_API_KEY")))))
	mux.Handle("OPTIONS /parse-recipe", preflight)

	// Meal plan endpoints
	mux.Handle("GET /meal-plans", protected(handleDesignGetAllMealPlans(recipeDatabase)))
	mux.Handle("POST /meal-plans", protected(handleDesignCreateMealPlan(recipeDatabase)))
	mux.Handle("DELETE /meal-plans/{meal_plan_id}", protected(handleDesignDeleteMealPlan(recipeDatabase)))
	mux.Handle("OPTIONS /meal-plans", preflight)
	mux.Handle("OPTIONS /meal-plans/{meal_plan_id}", preflight)

	// Auth endpoints
	mux.Handle("GET /auth/me", cors(handleGetUserInfo(authDatabase)))
	mux.Handle("POST /auth/register", cors(handlePasswordRegister(authDatabase)))
	mux.Handle("POST /auth/login", cors(handlePasswordLogin(authDatabase)))
	mux.Handle("OPTIONS /auth/me", preflight)
	mux.Handle("OPTIONS /auth/register", preflight)
	mux.Handle("OPTIONS /auth/login", preflight)

	// Google OAuth endpoints
	mux.Handle("GET /auth/google", cors(handleGoogleLogin(googleCfg)))
	mux.Handle("GET /auth/google/callback", cors(handleGoogleCallback(googleCfg, googleDB)))
	mux.Handle("OPTIONS /auth/google", preflight)
	mux.Handle("OPTIONS /auth/google/callback", preflight)

	// Grocery list endpoints
	mux.Handle("GET /grocery-list/{meal_plan_id}", protected(handleDesignGetGroceryList(recipeDatabase)))
	mux.Handle("POST /grocery-lists", protected(handleDesignCreateGroceryList(recipeDatabase)))
	mux.Handle("GET /grocery-lists", protected(handleDesignGetAllGroceryLists(recipeDatabase)))
	mux.Handle("GET /grocery-lists/{id}", protected(handleDesignGetGroceryListByID(recipeDatabase)))
	mux.Handle("PUT /grocery-lists/{id}", protected(handleDesignUpdateGroceryList(recipeDatabase)))
	mux.Handle("DELETE /grocery-lists/{id}", protected(handleDesignDeleteGroceryList(recipeDatabase)))
	mux.Handle("OPTIONS /grocery-lists", preflight)
	mux.Handle("OPTIONS /grocery-lists/{id}", preflight)
}

// setupSPAHandler configures serving the frontend SPA
// Serves static files from dist/ and falls back to index.html for client-side routing
func setupSPAHandler(mux *http.ServeMux) {
	ex, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	exeDir := filepath.Dir(ex)
	distDir := filepath.Join(exeDir, "dist")

	// Custom SPA file server that serves index.html for non-existent files
	spaHandler := func(w http.ResponseWriter, r *http.Request) {
		// Check if the requested file exists
		filePath := filepath.Join(distDir, r.URL.Path)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// File doesn't exist, serve index.html for SPA routing
			http.ServeFile(w, r, filepath.Join(distDir, "index.html"))
			return
		}
		// File exists, serve it normally
		http.FileServer(http.Dir(distDir)).ServeHTTP(w, r)
	}

	mux.Handle("/", http.HandlerFunc(spaHandler))
}
