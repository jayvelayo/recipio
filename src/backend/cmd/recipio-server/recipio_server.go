package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	rec "github.com/jayvelayo/recipio/internal/recipes"
	"github.com/jayvelayo/recipio/internal/sqlite_db"
)

const debugLogPath = "/Users/jayvee/repos/recipio/debug-95486f.log"

func debugLog(location, message string, data map[string]interface{}, hypothesisId string) {
	payload, _ := json.Marshal(map[string]interface{}{
		"sessionId": "95486f", "location": location, "message": message, "data": data,
		"timestamp": time.Now().UnixMilli(), "hypothesisId": hypothesisId,
	})
	f, err := os.OpenFile(debugLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	f.Write(payload)
	f.Write([]byte("\n"))
	f.Close()
}

// Design API types (doc/server_design.md)
type designRecipeRequest struct {
	Name        string   `json:"name"`
	Ingredients []string `json:"ingredients"`
	Steps       []string `json:"steps"`
}

type designCreateRecipeResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type designRecipeResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Ingredients []string `json:"ingredients"`
	Steps       []string `json:"steps"`
}

type designCreateMealPlanRequest struct {
	Recipes []string `json:"recipes"`
}

type designCreateMealPlanResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type designGroceryListResponse struct {
	Ingredients []string `json:"ingredients"`
}

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

func designRecipeToInternal(body designRecipeRequest) rec.Recipe {
	var recipe rec.Recipe
	recipe.Name = body.Name
	recipe.Instructions = rec.InstructionList(body.Steps)
	for _, line := range body.Ingredients {
		words := strings.Fields(line)
		if len(words) == 0 {
			continue
		}
		recipe.Ingredients = append(recipe.Ingredients, rec.Ingredient{
			Name:     words[len(words)-1],
			Quantity: strings.Join(words[0:len(words)-1], " "),
		})
	}
	return recipe
}

func internalRecipeToDesign(recipe rec.Recipe) designRecipeResponse {
	idStr := strconv.Itoa(recipe.ID)
	ingStrings := make([]string, 0, len(recipe.Ingredients))
	for _, ing := range recipe.Ingredients {
		s := strings.TrimSpace(ing.Quantity) + " " + strings.TrimSpace(ing.Name)
		ingStrings = append(ingStrings, strings.TrimSpace(s))
	}
	return designRecipeResponse{
		ID:          idStr,
		Name:        recipe.Name,
		Ingredients: ingStrings,
		Steps:       recipe.Instructions,
	}
}

func handleDesignCreateRecipe(recipeDb rec.RecipeDatabase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}
		var body designRecipeRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}
		recipe := designRecipeToInternal(body)
		if recipe.Name == "" || len(recipe.Ingredients) == 0 || len(recipe.Instructions) == 0 {
			http.Error(w, "Missing required fields: name, ingredients, steps", http.StatusBadRequest)
			return
		}
		recipeID, err := recipeDb.CreateRecipe(recipe)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		encodeJson(w, http.StatusCreated, designCreateRecipeResponse{
			ID:      strconv.FormatUint(recipeID, 10),
			Message: "Recipe created successfully",
		})
	})
}

func handleDesignGetRecipe(recipeDb rec.RecipeDatabase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil || id < 1 {
			http.Error(w, "Invalid recipe id", http.StatusBadRequest)
			return
		}
		recipe, err := recipeDb.GetRecipe(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		encodeJson(w, http.StatusOK, internalRecipeToDesign(recipe))
	})
}

func handleDesignGetAllRecipes(recipeDb rec.RecipeDatabase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recipes, err := recipeDb.GetAllRecipes()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list := make([]designRecipeResponse, 0, len(recipes))
		for _, recipe := range recipes {
			list = append(list, internalRecipeToDesign(recipe))
		}
		encodeJson(w, http.StatusOK, list)
	})
}

func handleDesignGetAllMealPlans(recipeDb rec.RecipeDatabase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// #region agent log
		debugLog("recipio_server.go:handleDesignGetAllMealPlans", "GET meal-plans received", map[string]interface{}{"path": r.URL.Path}, "H4")
		// #endregion
		plans, err := recipeDb.GetAllMealPlans()
		if err != nil {
			// #region agent log
			debugLog("recipio_server.go:handleDesignGetAllMealPlans", "GetAllMealPlans error", map[string]interface{}{"err": err.Error()}, "H3")
			// #endregion
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// #region agent log
		debugLog("recipio_server.go:handleDesignGetAllMealPlans", "sending response", map[string]interface{}{"planCount": len(plans)}, "H3")
		// #endregion
		encodeJson(w, http.StatusOK, plans)
	})
}

func handleDesignCreateMealPlan(recipeDb rec.RecipeDatabase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}
		var body designCreateMealPlanRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}
		mealPlanID, err := recipeDb.CreateMealPlan(body.Recipes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		encodeJson(w, http.StatusCreated, designCreateMealPlanResponse{
			ID:      mealPlanID,
			Message: "Meal plan created successfully",
		})
	})
}

func handleDesignGetGroceryList(recipeDb rec.RecipeDatabase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mealPlanID := r.PathValue("meal_plan_id")
		if mealPlanID == "" {
			http.Error(w, "Missing meal plan id", http.StatusBadRequest)
			return
		}
		ingredients, err := recipeDb.GetGroceryList(mealPlanID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		encodeJson(w, http.StatusOK, designGroceryListResponse{Ingredients: ingredients})
	})
}

func handleDesignDeleteMealPlan(recipeDb rec.RecipeDatabase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mealPlanID := r.PathValue("meal_plan_id")
		if mealPlanID == "" {
			http.Error(w, "Missing meal plan id", http.StatusBadRequest)
			return
		}
		err := recipeDb.DeleteMealPlan(mealPlanID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

type designCreateGroceryListRequest struct {
	Name       string                `json:"name"`
	Items      []rec.GroceryListItem `json:"items"`
	MealPlanID *string               `json:"meal_plan_id,omitempty"`
}

type designCreateGroceryListResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func handleDesignCreateGroceryList(recipeDb rec.RecipeDatabase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}
		var body designCreateGroceryListRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}
		if body.Name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}
		id, err := recipeDb.CreateGroceryList(body.Name, body.Items, body.MealPlanID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		encodeJson(w, http.StatusCreated, designCreateGroceryListResponse{ID: id, Name: body.Name})
	})
}

func handleDesignGetAllGroceryLists(recipeDb rec.RecipeDatabase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lists, err := recipeDb.GetAllGroceryLists()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		encodeJson(w, http.StatusOK, lists)
	})
}

func handleDesignGetGroceryListByID(recipeDb rec.RecipeDatabase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing grocery list id", http.StatusBadRequest)
			return
		}
		list, err := recipeDb.GetGroceryListByID(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		encodeJson(w, http.StatusOK, list)
	})
}

func handleDesignUpdateGroceryList(recipeDb rec.RecipeDatabase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing grocery list id", http.StatusBadRequest)
			return
		}
		var items []rec.GroceryListItem
		if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}
		err := recipeDb.UpdateGroceryList(id, items)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

func handleDesignDeleteGroceryList(recipeDb rec.RecipeDatabase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing grocery list id", http.StatusBadRequest)
			return
		}
		err := recipeDb.DeleteGroceryList(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

func withCORS(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		// Allow localhost and 127.0.0.1 origins (any port) so Vite and other dev servers work
		corsAllowed := strings.HasPrefix(origin, "http://127.0.0.1") || strings.HasPrefix(origin, "https://127.0.0.1") ||
			strings.HasPrefix(origin, "http://localhost") || strings.HasPrefix(origin, "https://localhost")
		// #region agent log
		log.Println("recipio_server.go:withCORS", "CORS check", map[string]interface{}{"origin": origin, "corsAllowed": corsAllowed, "path": r.URL.Path}, "H1")
		// #endregion
		if corsAllowed {
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

func SetUpRoutes(
	mux *http.ServeMux,
	recipeDatabase rec.RecipeDatabase,
) {
	// Design API (doc/server_design.md)
	mux.Handle("POST /recipes", withCORS(handleDesignCreateRecipe(recipeDatabase)))
	mux.Handle("GET /recipes/{id}", withCORS(handleDesignGetRecipe(recipeDatabase)))
	mux.Handle("GET /recipes", withCORS(handleDesignGetAllRecipes(recipeDatabase)))
	mux.Handle("OPTIONS /recipes", withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))
	mux.Handle("GET /meal-plans", withCORS(handleDesignGetAllMealPlans(recipeDatabase)))
	mux.Handle("POST /meal-plans", withCORS(handleDesignCreateMealPlan(recipeDatabase)))
	mux.Handle("OPTIONS /meal-plans", withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))
	mux.Handle("DELETE /meal-plans/{meal_plan_id}", withCORS(handleDesignDeleteMealPlan(recipeDatabase)))
	mux.Handle("OPTIONS /meal-plans/{meal_plan_id}", withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))
	mux.Handle("GET /grocery-list/{meal_plan_id}", withCORS(handleDesignGetGroceryList(recipeDatabase)))
	mux.Handle("POST /grocery-lists", withCORS(handleDesignCreateGroceryList(recipeDatabase)))
	mux.Handle("GET /grocery-lists", withCORS(handleDesignGetAllGroceryLists(recipeDatabase)))
	mux.Handle("GET /grocery-lists/{id}", withCORS(handleDesignGetGroceryListByID(recipeDatabase)))
	mux.Handle("PUT /grocery-lists/{id}", withCORS(handleDesignUpdateGroceryList(recipeDatabase)))
	mux.Handle("DELETE /grocery-lists/{id}", withCORS(handleDesignDeleteGroceryList(recipeDatabase)))
	mux.Handle("OPTIONS /grocery-lists", withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))
	mux.Handle("OPTIONS /grocery-lists/{id}", withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))
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
		log.Fatalf("unable to init db: %v", err)
	}
	defer recipeDb.CloseDb()
	srv := newServer(recipeDb)

	log.Println("Starting server on :4002...")
	if err := http.ListenAndServe(":4002", srv); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
