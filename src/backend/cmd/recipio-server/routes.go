package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	rec "github.com/jayvelayo/recipio/internal/recipes"
)

// SetUpRoutes registers all API endpoints and static file serving
// This is the single place to see the entire API contract
func SetUpRoutes(
	mux *http.ServeMux,
	recipeDatabase rec.RecipeDatabase,
	allowedOrigins []string,
) {
	// ============================================================
	// STATIC FILE SERVING
	// ============================================================
	setupSPAHandler(mux)

	// ============================================================
	// RECIPE ENDPOINTS (Design API)
	// ============================================================
	mux.Handle("POST /recipes", withCORS(allowedOrigins, handleDesignCreateRecipe(recipeDatabase)))
	mux.Handle("GET /recipes/{id}", withCORS(allowedOrigins, handleDesignGetRecipe(recipeDatabase)))
	mux.Handle("GET /recipes", withCORS(allowedOrigins, handleDesignGetAllRecipes(recipeDatabase)))
	mux.Handle("DELETE /recipes/{id}", withCORS(allowedOrigins, handleDesignDeleteRecipe(recipeDatabase)))
	mux.Handle("OPTIONS /recipes", withCORS(allowedOrigins, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))

	// ============================================================
	// AI RECIPE PARSING ENDPOINT
	// ============================================================
	mux.Handle("POST /parse-recipe", withCORS(allowedOrigins, handleDesignParseRecipe()))
	mux.Handle("OPTIONS /parse-recipe", withCORS(allowedOrigins, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))

	// ============================================================
	// MEAL PLAN ENDPOINTS (Design API)
	// ============================================================
	mux.Handle("GET /meal-plans", withCORS(allowedOrigins, handleDesignGetAllMealPlans(recipeDatabase)))
	mux.Handle("POST /meal-plans", withCORS(allowedOrigins, handleDesignCreateMealPlan(recipeDatabase)))
	mux.Handle("DELETE /meal-plans/{meal_plan_id}", withCORS(allowedOrigins, handleDesignDeleteMealPlan(recipeDatabase)))
	mux.Handle("OPTIONS /meal-plans", withCORS(allowedOrigins, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))
	mux.Handle("OPTIONS /meal-plans/{meal_plan_id}", withCORS(allowedOrigins, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))

	// ============================================================
	// GROCERY LIST ENDPOINTS (Design API)
	// ============================================================
	mux.Handle("GET /grocery-list/{meal_plan_id}", withCORS(allowedOrigins, handleDesignGetGroceryList(recipeDatabase)))
	mux.Handle("POST /grocery-lists", withCORS(allowedOrigins, handleDesignCreateGroceryList(recipeDatabase)))
	mux.Handle("GET /grocery-lists", withCORS(allowedOrigins, handleDesignGetAllGroceryLists(recipeDatabase)))
	mux.Handle("GET /grocery-lists/{id}", withCORS(allowedOrigins, handleDesignGetGroceryListByID(recipeDatabase)))
	mux.Handle("PUT /grocery-lists/{id}", withCORS(allowedOrigins, handleDesignUpdateGroceryList(recipeDatabase)))
	mux.Handle("DELETE /grocery-lists/{id}", withCORS(allowedOrigins, handleDesignDeleteGroceryList(recipeDatabase)))
	mux.Handle("OPTIONS /grocery-lists", withCORS(allowedOrigins, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))
	mux.Handle("OPTIONS /grocery-lists/{id}", withCORS(allowedOrigins, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))
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
