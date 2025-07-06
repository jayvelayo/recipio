package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	rec "github.com/jayvelayo/recipio/internal/recipes"
	"github.com/jayvelayo/recipio/internal/sqlite_db"
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

type ResponseStatus int

const (
	StatusOK ResponseStatus = iota
	StatusError
	StatusEmptyBody
	StatusInvalidJson
	StatusNotFound
	StatusMissingFields
	StatusAlreadyExist
	StatusEncodingError
)

type RecipeBody struct {
	Name         string   `json:"name"`
	Ingredients  []string `json:"ingredients"`
	Instructions []string `json:"instructions"`
}

type CreateRecipeResponse struct {
	Status       ResponseStatus `json:"status"`
	ErrorMessage string         `json:"errorMessage"`
	RecipeId     uint64         `json:"id"`
}

func returnError(w http.ResponseWriter, httpStatus int, status ResponseStatus, errorMessage string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	errReponse := CreateRecipeResponse{
		Status:       status,
		ErrorMessage: errorMessage,
		RecipeId:     0,
	}
	json.NewEncoder(w).Encode(errReponse)
}

func handleCreateRecipe(recipeDb rec.RecipeDatabase) http.Handler {
	convertBodyToRecipe := func(body RecipeBody) rec.Recipe {
		var recipe rec.Recipe
		recipe.Name = body.Name
		recipe.Instructions = body.Instructions
		for _, line := range body.Ingredients {
			words := strings.Fields(line)
			if len(words) == 0 {
				continue
			}
			ingredients := rec.Ingredient{
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
				returnError(w, http.StatusBadRequest, StatusEmptyBody, "Request body cannot be empty")
				return
			}
			recipeBody, err := decodeJson[RecipeBody](r)
			if err != nil {
				returnError(w, http.StatusBadRequest, StatusInvalidJson, "Invalid JSON body")
				return
			}
			recipe := convertBodyToRecipe(recipeBody)
			if recipe.Name == "" || len(recipe.Ingredients) == 0 || len(recipe.Instructions) == 0 {
				returnError(w, http.StatusBadRequest, StatusMissingFields, "Missing fields")
				return
			}
			recipeId, err := recipeDb.CreateRecipe(recipe)
			if err != nil {
				returnError(w, http.StatusBadRequest, StatusError, err.Error())
				return
			}
			err = encodeJson(w, http.StatusCreated, CreateRecipeResponse{RecipeId: recipeId})
			if err != nil {
				returnError(w, http.StatusBadRequest, StatusEncodingError, err.Error())
				return
			}
		},
	)
}

func handleGetRecipe(recipeDb rec.RecipeDatabase) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			id_str := r.PathValue("id")
			id, err := strconv.Atoi(id_str)
			if err != nil || id < 1 {
				http.Error(w, "Failed to parse recipe id", http.StatusBadRequest)
				return
			}
			var recipes rec.Recipes
			log.Printf("Looking for recipe id %d", id)
			if id != 0 {
				recipe, err := recipeDb.GetRecipe(id)
				if err != nil {
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}
				recipes = rec.Recipes{recipe}
			}
			err = encodeJson(w, http.StatusOK, recipes)
			if err != nil {
				http.Error(w, "Error marshalling recipe to JSON", http.StatusInternalServerError)
				return
			}
		},
	)
}

func handleGetAllRecipe(recipeDb rec.RecipeDatabase) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var recipes rec.Recipes
			recipes, err := recipeDb.GetAllRecipes()
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

func handleDeleteRecipe(recipeDb rec.RecipeDatabase) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			id_str := r.PathValue("id")
			id, err := strconv.Atoi(id_str)
			if err != nil || id < 1 {
				http.Error(w, "Failed to parse recipe id", http.StatusBadRequest)
				return
			}
			log.Printf("Deleting recipe id %d", id)
			if id != 0 {
				err := recipeDb.DeleteRecipe(id)
				if err != nil {
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}
			}
			w.WriteHeader(http.StatusOK)
		},
	)
}

func withCORS(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
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
	recipeDatabase rec.RecipeDatabase,
) {
	mux.Handle("OPTIONS /v1/recipe", handleCORS())
	mux.Handle("GET /v1/recipe", withCORS(handleGetAllRecipe(recipeDatabase)))
	mux.Handle("POST /v1/recipe", withCORS(handleCreateRecipe(recipeDatabase)))
	mux.Handle("GET /v1/recipe/{id}", withCORS(handleGetRecipe(recipeDatabase)))
	mux.Handle("DELETE /v1/recipe/{id}", withCORS(handleDeleteRecipe(recipeDatabase)))
	mux.Handle("OPTIONS /v1/recipe/{id}", handleCORS())
}

func newServer(
	recipeDatabase rec.RecipeDatabase,
) http.Handler {
	mux := http.NewServeMux()
	SetUpRoutes(
		mux,
		recipeDatabase,
	)
	return mux
}

func main() {
	recipeDb, err := sqlite_db.InitDb("recipes.db")
	if err != nil {
		log.Fatalf("unable to init db")
	}
	defer recipeDb.CloseDb()
	srv := newServer(recipeDb)

	log.Println("Starting server on :4002...")
	if err := http.ListenAndServe(":4002", srv); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
