package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
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

func encodeJson[T any](w http.ResponseWriter, r *http.Request, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
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
			err = CreateRecipe(recipe, recipeDb)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("Recipe created successfully!"))
		},
	)
}

/*
	returns the id and false. id of 0 means 'all'
	If the path is not specified, or invalid, then returns true.
*/
func getIdFromPath(r *http.Request) (int, bool) {
	paths := strings.Split(r.URL.Path, "/")
	// we expect v1/recipe/{id}
	if paths[0] != "v1" || paths[1] != "recipe" {
		return 0, true
	}
	if len(paths) < 3 {
		return 0, false
	}
	if len(paths) > 3 {
		return 0, true
	}
	id, err := strconv.Atoi(paths[2])
	if err != nil || id < 1 {
		return 0, true
	}
	return id, false
}

func handleGetRecipe(recipeDb RecipeDatabase) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			id, is_err := getIdFromPath(r)
			if is_err {
				http.Error(w, "Invalid recipe path", http.StatusBadRequest)
				return
			}
			var recipes Recipes
			if id != 0 {
				recipe, err := GetRecipe(id, recipeDb)
				if err != nil {
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}
				recipes = Recipes{recipe}
			} else {
				recipes_, err := FetchAllRecipes(recipeDb)
				if err != nil {
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}
				recipes = recipes_
			}
			err := encodeJson(w, r, http.StatusOK, recipes)
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
