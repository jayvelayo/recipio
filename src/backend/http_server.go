package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func decodeJson[T any](r *http.Request) (T, error) {
	var v T
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&v)
	if err != nil {
		return v, fmt.Errorf("json decode error %w", err)
	}
	return v, err
}

func encodeJson[T any](w http.ResponseWriter, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(v)
	return nil
}

type CreateRecipeResponse struct {
	RecipeId uint64 `json:"id"`
}

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
			recipe, err := decodeJson[Recipe](r)
			if err != nil {
				http.Error(w, "Invalid JSON body", http.StatusBadRequest)
				return
			}
			if recipe.Name == "" || len(recipe.Ingredients) == 0 || len(recipe.Instructions) == 0 {
				http.Error(w, "Missing fields", http.StatusBadRequest)
				return
			}
			recipeId, err := recipeDb.createRecipe(recipe)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			err = encodeJson(w, http.StatusCreated, CreateRecipeResponse{RecipeId: recipeId})
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
		},
	)
}

func handleGetRecipe(recipeDb RecipeDatabase) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			id_str := r.PathValue("id")
			id, err := strconv.Atoi(id_str)
			if err != nil || id < 1 {
				http.Error(w, "Failed to parse recipe id", http.StatusBadRequest)
				return
			}
			var recipes Recipes
			if id != 0 {
				recipe, err := recipeDb.getRecipe(id)
				if err != nil {
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}
				recipes = Recipes{recipe}
			}
			err = encodeJson(w, http.StatusOK, recipes)
			if err != nil {
				http.Error(w, "Error marshalling recipe to JSON", http.StatusInternalServerError)
				return
			}
		},
	)
}

func handleGetAllRecipe(recipeDb RecipeDatabase) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var recipes Recipes
			recipes, err := recipeDb.getAllRecipes()
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}

			err = encodeJson(w, http.StatusOK, recipes)
			if err != nil {
				http.Error(w, "Error marshalling recipe to JSON", http.StatusInternalServerError)
				return
			}
		},
	)
}

func SetUpRoutes(
	mux *http.ServeMux,
	recipeDatabase RecipeDatabase,
) {
	mux.Handle("POST /v1/recipe", handleCreateRecipe(recipeDatabase))
	mux.Handle("GET /v1/recipe/{id}", handleGetRecipe(recipeDatabase))
	mux.Handle("GET /v1/recipe", handleGetAllRecipe(recipeDatabase))
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
	recipeDb, err := initDb()
	if err != nil {
		log.Fatalf("unable to init db")
	}
	defer recipeDb.closeDb()
	srv := newServer(recipeDb)

	log.Println("Starting server on :4002...")
	if err := http.ListenAndServe(":4002", srv); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
