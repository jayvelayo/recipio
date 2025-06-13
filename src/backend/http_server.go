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

func encodeJson[T any](w http.ResponseWriter, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	fmt.Println("Response Headers:")
	for name, values := range w.Header() {
		for _, value := range values {
			fmt.Printf("%s: %s\n", name, value)
		}
	}
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(v)
	return nil
}

type CreateRecipeResponse struct {
	RecipeId uint64 `json:"id"`
}

func handleCreateRecipe(recipeDb RecipeDatabase) http.Handler {

	type RecipeBody struct {
		Name         string   `json:"name"`
		Ingredients  []string `json:"ingredients"`
		Instructions []string `json:"instructions"`
	}

	convertBodyToRecipe := func(body RecipeBody) Recipe {
		var recipe Recipe
		recipe.Name = body.Name
		recipe.Instructions = body.Instructions
		for _, line := range body.Ingredients {
			words := strings.Fields(line)
			if len(words) == 0 {
				continue
			}
			ingredients := Ingredient{
				Name:     words[len(words)-1],
				Quantity: strings.Join(words[0:len(words)-1], " "),
			}
			recipe.Ingredients = append(recipe.Ingredients, ingredients)
		}
		return recipe
	}

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
			recipeBody, err := decodeJson[RecipeBody](r)
			if err != nil {
				http.Error(w, "Invalid JSON body", http.StatusBadRequest)
				return
			}
			log.Printf("Received body: %+v", recipeBody)
			recipe := convertBodyToRecipe(recipeBody)
			log.Printf("Processed body: %+v", recipe)
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
			log.Printf("Looking for recipe id %d", id)
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
			log.Println("Called `handleGetAllRecipe`")
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

func withCORS(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		log.Printf("Received request from %s", origin)
		// Allow all localhost origins (any port)
		if strings.HasPrefix(origin, "http://127.0.0.1") || strings.HasPrefix(origin, "https://127.0.0.1") {
			log.Printf("Setting headers")
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin") // Important: informs caches that response varies by Origin
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

func handleCORS() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		log.Printf("Received OPTIONS request from %s", origin)
		if strings.HasPrefix(origin, "http://127.0.0.1") || strings.HasPrefix(origin, "https://127.0.0.1") {
			log.Printf("Setting headers")
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin") // Important: informs caches that response varies by Origin
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func SetUpRoutes(
	mux *http.ServeMux,
	recipeDatabase RecipeDatabase,
) {
	mux.Handle("OPTIONS /v1/recipe", handleCORS())
	mux.Handle("POST /v1/recipe", withCORS(handleCreateRecipe(recipeDatabase)))
	mux.Handle("GET /v1/recipe/{id}", withCORS(handleGetRecipe(recipeDatabase)))
	mux.Handle("GET /v1/recipe", withCORS(handleGetAllRecipe(recipeDatabase)))
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
