package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func handleCreateRecipe(recipeDb RecipeDatabase) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var err error
			if r.Header.Get("Content-Type") != "application/json" {
				http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			}
			if r.Body == nil || r.Body == http.NoBody {
				http.Error(w, "Request body cannot be empty", http.StatusBadRequest)
				return
			}
			decoder := json.NewDecoder(r.Body)

			var recipe Recipe
			err = decoder.Decode(&recipe)
			if err != nil {
				http.Error(w, "Invalid JSON body", http.StatusBadRequest)
				return
			}

			err = CreateRecipe(recipe, recipeDb) // Replace nil with actual database instance
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("Recipe created successfully!"))
		},
	)
}

func handleGetRecipe(recipeDb RecipeDatabase) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			idParam := r.URL.Query().Get("id")
			if idParam == "" {
				http.Error(w, "Missing recipe ID", http.StatusBadRequest)
				return
			}
			id, err := strconv.Atoi(idParam)
			if err != nil || id < 1 {
				http.Error(w, "Invalid recipe ID", http.StatusBadRequest)
				return
			}
			recipe, err := GetRecipe(id, recipeDb)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			jsonData, err := json.Marshal(recipe)
			if err != nil {
				http.Error(w, "Error marshalling recipe to JSON", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(jsonData)
		},
	)
}

func SetUpRoutes(
	mux *http.ServeMux,
	recipeDatabase RecipeDatabase,
) {
	mux.Handle("POST /v1/recipe", handleCreateRecipe(recipeDatabase))
	mux.Handle("GET /v1/recipe", handleGetRecipe(recipeDatabase))

}

func newServer(
	recipeDatabase RecipeDatabase,
) http.Handler {
	mux := http.NewServeMux()
	SetUpRoutes(
		mux,
		recipeDatabase,
	)
	return mux
}

func main() {
	var recipeDb RecipeDatabase = &LocalRecipeDatabase{}
	srv := newServer(recipeDb)

	log.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", srv); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
