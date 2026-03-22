package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	rec "github.com/jayvelayo/recipio/internal/recipes"
	"github.com/jayvelayo/recipio/internal/sqlite_db"
)

func createRequestWithBody(method, url string, body interface{}) (*http.Request, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func getResponseBody(r *httptest.ResponseRecorder) string {
	bodybytes, err := io.ReadAll(r.Body)
	if err != nil {
		return "unable to read response"
	}
	return (string(bodybytes))
}

func createFakeServer(db rec.RecipeDatabase) http.Handler {
	mux := http.NewServeMux()
	SetUpRoutes(mux, db)
	return mux
}

/*
func TestCreateRecipeHandler(t *testing.T) {
	handler := createFakeServer(&rec.MockRecipeDatabase{})
	t.Run("Testing that endpoint must accept json", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/v1/recipe", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusUnsupportedMediaType {
			t.Errorf("Expected status code %d, got %d", http.StatusUnsupportedMediaType, response.Code)
		}
	})

	t.Run("Testing that json body has the name keys", func(t *testing.T) {
		incorrectBody := rec.Recipe{
			Description: "A delicious recipe",
		}
		req, _ := createRequestWithBody("POST", "/v1/recipe", incorrectBody)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, response.Code)
		}
	})

	t.Run("Creates a new recipe", func(t *testing.T) {
		correctBody := RecipeBody{
			Name: "My Recipe",
			Ingredients: []string{
				"2 Banana",
				"2 Egg",
			},
			Instructions: []string{"Cook", "me"},
		}
		req, _ := createRequestWithBody("POST", "/v1/recipe", correctBody)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, req)
		if response.Code != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d", http.StatusCreated, response.Code)
		}
	})
}


func TestGetRecipeHandler(t *testing.T) {

	mockdb := rec.MockRecipeDatabase{}
	mockdb.recipes = append(mockdb.recipes, Recipe{ID: 1}, Recipe{ID: 2})
	mockdb.recipeCount = 2
	handler := createFakeServer(&mockdb)
	t.Run("Fetches a recipe by ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/recipe/1", nil)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, req)
		if response.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, response.Code)
		}
	})

	t.Run("Fetches all recipes", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/recipe", nil)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, req)
		if response.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, response.Code)
		}
	})
	t.Run("Checks if id is found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/recipe/3", nil)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, req)
		if response.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, response.Code)
		}
	})
}
*/

var mockRecipes = []RecipeBody{
	{
		Name: "French toast",
		Ingredients: []string{
			"2 pieces eggs",
			"2 pieces bread",
			"50 mL milk",
		},
		Instructions: []string{
			"Beats eggs until smooth",
			"Add milk to the egg mixture",
			"Dip the bread into the milk-egg mixture",
			"Cook on a non-stick pan with butter until desired texture",
		},
	},
	{
		Name: "Pancakes",
		Ingredients: []string{
			"200 grams flour",
			"300 mL milk",
			"1 pieces egg",
			"2 tablespoons sugar",
			"1 teaspoons baking powder",
		},
		Instructions: []string{
			"Mix all dry ingredients in a bowl",
			"Add milk and egg, then whisk until smooth",
			"Heat a lightly oiled pan over medium heat",
			"Pour batter onto the pan and cook until bubbles form, then flip and cook until golden",
		},
	},
	{
		Name: "Scrambled Eggs",
		Ingredients: []string{
			"3 pieces eggs",
			"30 mL milk",
			"1 tablespoons butter",
			"0.5 teaspoons salt",
		},
		Instructions: []string{
			"Crack the eggs into a bowl and add milk and salt",
			"Whisk the mixture until well combined",
			"Melt butter in a pan over medium heat",
			"Pour in the egg mixture and stir gently until just set",
		},
	},
	{
		Name: "Grilled Cheese Sandwich",
		Ingredients: []string{
			"2 pieces bread",
			"2 pieces cheese slices",
			"1 tablespoons butter",
		},
		Instructions: []string{
			"Butter one side of each bread slice",
			"Place cheese between the unbuttered sides of the bread",
			"Cook in a pan over medium heat until golden on both sides and cheese is melted",
		},
	},
	{
		Name: "Banana Smoothie",
		Ingredients: []string{
			"1 pieces banana",
			"200 mL milk",
			"100 grams yogurt",
			"1 tablespoons honey",
		},
		Instructions: []string{
			"Peel and slice the banana",
			"Add banana, milk, yogurt, and honey to a blender",
			"Blend until smooth",
			"Serve chilled",
		},
	},
}

func initTestDb() (rec.RecipeDatabase, error) {
	test_db_name := "test_recipes.db"
	os.Remove(test_db_name)
	return sqlite_db.InitDb(test_db_name)
}

func decodeJsonResponse[T any](r *bytes.Buffer) (T, error) {
	var v T
	err := json.NewDecoder(r).Decode(&v)
	if err != nil {
		return v, fmt.Errorf("json decode error %w", err)
	}
	return v, err
}

func TestServer_e2e_recipes(t *testing.T) {
	testDb, err := initTestDb()
	if err != nil {
		t.Fatalf("unable to initialize test db %v", err)
	}
	defer testDb.CloseDb()
	handler := createFakeServer(testDb)
	t.Run("Creates a new recipe", func(t *testing.T) {
		correctBody := mockRecipes[0]
		req, _ := createRequestWithBody("POST", "/v1/recipe", correctBody)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, req)

		if response.Code != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d", http.StatusCreated, response.Code)
		}
		recipeResp, err := decodeJsonResponse[CreateRecipeResponse](response.Body)
		if err != nil {
			t.Fatalf("failed to decode response body: %s", err)
		}
		if recipeResp.Status != StatusOK {
			t.Errorf("Expected status %d, got %d", StatusOK, recipeResp.Status)
		}
		if recipeResp.RecipeId == 0 {
			t.Log(recipeResp)
			t.Errorf("Expected id to be non-zero: %d", recipeResp.RecipeId)
		}
	})

	t.Run("Fetches a recipe by ID", func(t *testing.T) {
		/* insert first */
		expected_recipe := mockRecipes[1]
		req, _ := createRequestWithBody("POST", "/v1/recipe", expected_recipe)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, req)
		if response.Code != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d", http.StatusCreated, response.Code)
		}
		recipeResp, err := decodeJsonResponse[CreateRecipeResponse](response.Body)
		if err != nil {
			t.Fatalf("failed to decode response body: %s", err)
		}
		if recipeResp.RecipeId == 0 {
			t.Log(recipeResp)
			t.Errorf("Expected id to be non-zero: %d", recipeResp.RecipeId)
		}

		/* fetch */
		fetch_endpoint := fmt.Sprintf("/v1/recipe/%d", recipeResp.RecipeId)
		req, _ = http.NewRequest("GET", fetch_endpoint, nil)
		response = httptest.NewRecorder()

		handler.ServeHTTP(response, req)
		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %d, got %d", http.StatusOK, response.Code)
		}
		recipes, err := decodeJsonResponse[rec.Recipes](response.Body)
		if err != nil {
			t.Fatalf("failed to decode response body: %s", err)
		}
		if len(recipes) != 1 {
			t.Errorf("Expected length of recipes arr %d, got %d", 1, len(recipes))
		}
		if recipes[0].ID == 0 {
			t.Errorf("Expected non-zero id, got %d", recipes[0].ID)
		}
	})
}

func TestServer_e2e_delete_recipe(t *testing.T) {
	testDb, err := initTestDb()
	if err != nil {
		t.Fatalf("unable to initialize test db %v", err)
	}
	defer testDb.CloseDb()
	handler := createFakeServer(testDb)
	t.Run("Deletes the recipe", func(t *testing.T) {
		/* insert first */
		expected_recipe := mockRecipes[1]
		req, _ := createRequestWithBody("POST", "/v1/recipe", expected_recipe)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, req)
		if response.Code != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d", http.StatusCreated, response.Code)
		}
		recipeResp, err := decodeJsonResponse[CreateRecipeResponse](response.Body)
		if err != nil {
			t.Fatalf("failed to decode response body: %s", err)
		}
		if recipeResp.RecipeId == 0 {
			t.Log(recipeResp)
			t.Errorf("Expected id to be non-zero: %d", recipeResp.RecipeId)
		}

		/* fetch */
		fetch_endpoint := fmt.Sprintf("/v1/recipe/%d", recipeResp.RecipeId)
		req, _ = http.NewRequest("GET", fetch_endpoint, nil)
		response = httptest.NewRecorder()

		handler.ServeHTTP(response, req)
		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %d, got %d", http.StatusOK, response.Code)
		}
		recipes, err := decodeJsonResponse[rec.Recipes](response.Body)
		if err != nil {
			t.Fatalf("failed to decode response body: %s", err)
		}
		if len(recipes) != 1 {
			t.Errorf("Expected length of recipes arr %d, got %d", 1, len(recipes))
		}
		if recipes[0].ID == 0 {
			t.Errorf("Expected non-zero id, got %d", recipes[0].ID)
		}

		/* delete */
		delete_endpoint := fmt.Sprintf("/v1/recipe/%d", recipeResp.RecipeId)
		req, _ = http.NewRequest("DELETE", delete_endpoint, nil)
		response = httptest.NewRecorder()

		handler.ServeHTTP(response, req)
		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %d, got %d", http.StatusOK, response.Code)
		}
	})
}

func TestServer_e2e_multiple_recipes(t *testing.T) {
	testDb, err := initTestDb()
	if err != nil {
		t.Fatalf("unable to initialize test db %v", err)
	}
	defer testDb.CloseDb()
	handler := createFakeServer(testDb)
	t.Run("Fetches multiple recipes", func(t *testing.T) {
		for _, recipe := range mockRecipes {
			req, _ := createRequestWithBody("POST", "/v1/recipe", recipe)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, req)
			if response.Code != http.StatusCreated {
				t.Errorf("Expected status code %d, got %d", http.StatusCreated, response.Code)
			}
		}
		/* fetch */
		req, _ := http.NewRequest("GET", "/v1/recipe", nil)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, req)
		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %d, got %d", http.StatusOK, response.Code)
		}
		recipes, err := decodeJsonResponse[rec.Recipes](response.Body)
		if err != nil {
			t.Fatalf("failed to decode response body: %s", err)
		}
		if len(recipes) != 5 {
			t.Errorf("Expected length of recipes arr %d, got %d", 5, len(recipes))
		}
		for _, recipe := range recipes {
			if recipe.ID == 0 {
				t.Errorf("Expected non-zero id, got %d", recipe.ID)
			}
		}
	})
}

// --- E2E tests per doc/server_design.md ---

// Design API types (design uses "steps" and specific response shapes)
type DesignRecipeRequest struct {
	Name        string   `json:"name"`
	Ingredients []string `json:"ingredients"`
	Steps       []string `json:"steps"`
}

type DesignCreateRecipeResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type DesignRecipeResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Ingredients []string `json:"ingredients"`
	Steps       []string `json:"steps"`
}

type DesignCreateMealPlanRequest struct {
	Recipes []string `json:"recipes"`
}

type DesignCreateMealPlanResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type DesignGroceryListResponse struct {
	Ingredients []string `json:"ingredients"`
}

type DesignMealPlanSummary struct {
	ID          string   `json:"id"`
	RecipeNames []string `json:"recipe_names"`
}

func TestServer_e2e_design_GET_meal_plans(t *testing.T) {
	testDb, err := initTestDb()
	if err != nil {
		t.Fatalf("unable to initialize test db: %v", err)
	}
	defer testDb.CloseDb()
	handler := createFakeServer(testDb)

	// Create two recipes
	var r1ID, r2ID string
	for _, name := range []string{"Pasta", "Salad"} {
		body := DesignRecipeRequest{Name: name, Ingredients: []string{"x"}, Steps: []string{"step"}}
		req, _ := createRequestWithBody("POST", "/recipes", body)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("setup create recipe: expected 201, got %d", rec.Code)
		}
		var res DesignCreateRecipeResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if name == "Pasta" {
			r1ID = res.ID
		} else {
			r2ID = res.ID
		}
	}

	// Create first meal plan (Pasta only)
	req, _ := createRequestWithBody("POST", "/meal-plans", DesignCreateMealPlanRequest{Recipes: []string{r1ID}})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create meal plan 1: expected 201, got %d body: %s", rec.Code, getResponseBody(rec))
	}
	var mp1 DesignCreateMealPlanResponse
	if err := json.NewDecoder(rec.Body).Decode(&mp1); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// Create second meal plan (Pasta + Salad)
	req, _ = createRequestWithBody("POST", "/meal-plans", DesignCreateMealPlanRequest{Recipes: []string{r1ID, r2ID}})
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create meal plan 2: expected 201, got %d body: %s", rec.Code, getResponseBody(rec))
	}

	// GET /meal-plans
	req, _ = http.NewRequest("GET", "/meal-plans", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("GET /meal-plans: expected 200, got %d body: %s", rec.Code, getResponseBody(rec))
	}
	var list []DesignMealPlanSummary
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatalf("decode GET /meal-plans: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("GET /meal-plans: expected 2 meal plans, got %d", len(list))
	}
	// First meal plan: one recipe (Pasta)
	if len(list[0].RecipeNames) != 1 || list[0].RecipeNames[0] != "Pasta" {
		t.Errorf("GET /meal-plans: first plan expected [Pasta], got %v", list[0].RecipeNames)
	}
	// Second meal plan: two recipes (Pasta, Salad)
	if len(list[1].RecipeNames) != 2 {
		t.Errorf("GET /meal-plans: second plan expected 2 recipe names, got %v", list[1].RecipeNames)
	}
	namesOk := (list[1].RecipeNames[0] == "Pasta" && list[1].RecipeNames[1] == "Salad") ||
		(list[1].RecipeNames[0] == "Salad" && list[1].RecipeNames[1] == "Pasta")
	if !namesOk {
		t.Errorf("GET /meal-plans: second plan expected Pasta and Salad, got %v", list[1].RecipeNames)
	}
	if list[0].ID == "" || list[1].ID == "" {
		t.Error("GET /meal-plans: expected non-empty id for each plan")
	}
}

func TestServer_e2e_design_POST_recipes(t *testing.T) {
	testDb, err := initTestDb()
	if err != nil {
		t.Fatalf("unable to initialize test db: %v", err)
	}
	defer testDb.CloseDb()
	handler := createFakeServer(testDb)

	body := DesignRecipeRequest{
		Name:        "Test Recipe",
		Ingredients: []string{"2 eggs", "1 cup flour"},
		Steps:       []string{"Mix ingredients", "Bake at 350F"},
	}
	req, err := createRequestWithBody("POST", "/recipes", body)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("POST /recipes: expected status %d, got %d body: %s", http.StatusCreated, rec.Code, getResponseBody(rec))
	}
	var res DesignCreateRecipeResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if res.ID == "" {
		t.Error("POST /recipes: expected non-empty id")
	}
	if res.Message != "Recipe created successfully" {
		t.Errorf("POST /recipes: expected message 'Recipe created successfully', got %q", res.Message)
	}
}

func TestServer_e2e_design_GET_recipes_id(t *testing.T) {
	testDb, err := initTestDb()
	if err != nil {
		t.Fatalf("unable to initialize test db: %v", err)
	}
	defer testDb.CloseDb()
	handler := createFakeServer(testDb)

	// Create a recipe first
	body := DesignRecipeRequest{
		Name:        "Fetch Me",
		Ingredients: []string{"1 apple", "2 bananas"},
		Steps:       []string{"Chop", "Serve"},
	}
	req, _ := createRequestWithBody("POST", "/recipes", body)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("setup create: expected 201, got %d", rec.Code)
	}
	var createRes DesignCreateRecipeResponse
	if err := json.NewDecoder(rec.Body).Decode(&createRes); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	// GET /recipes/{id}
	req, _ = http.NewRequest("GET", "/recipes/"+createRes.ID, nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("GET /recipes/{id}: expected 200, got %d body: %s", rec.Code, getResponseBody(rec))
	}
	var recipe DesignRecipeResponse
	if err := json.NewDecoder(rec.Body).Decode(&recipe); err != nil {
		t.Fatalf("decode GET response: %v", err)
	}
	if recipe.ID != createRes.ID {
		t.Errorf("GET /recipes/{id}: id mismatch: want %q got %q", createRes.ID, recipe.ID)
	}
	if recipe.Name != "Fetch Me" {
		t.Errorf("GET /recipes/{id}: name want 'Fetch Me', got %q", recipe.Name)
	}
	if len(recipe.Ingredients) != 2 || len(recipe.Steps) != 2 {
		t.Errorf("GET /recipes/{id}: ingredients len=%d steps len=%d", len(recipe.Ingredients), len(recipe.Steps))
	}
}

func TestServer_e2e_design_GET_recipes(t *testing.T) {
	testDb, err := initTestDb()
	if err != nil {
		t.Fatalf("unable to initialize test db: %v", err)
	}
	defer testDb.CloseDb()
	handler := createFakeServer(testDb)

	// Create two recipes
	for _, name := range []string{"First", "Second"} {
		body := DesignRecipeRequest{Name: name, Ingredients: []string{"x"}, Steps: []string{"step"}}
		req, _ := createRequestWithBody("POST", "/recipes", body)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("setup create: expected 201, got %d", rec.Code)
		}
	}

	req, _ := http.NewRequest("GET", "/recipes", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("GET /recipes: expected 200, got %d body: %s", rec.Code, getResponseBody(rec))
	}
	var list []DesignRecipeResponse
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatalf("decode GET /recipes: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("GET /recipes: expected 2 recipes, got %d", len(list))
	}
	// @TODO should check both recipes actual data (names, ingredients, steps, etc)
	// @TODO these kinds of tests are suited to be separated into t.Run blocks
	// if common setup, and multiple assertions are needed.
}

func TestServer_e2e_design_POST_meal_plans(t *testing.T) {
	testDb, err := initTestDb()
	if err != nil {
		t.Fatalf("unable to initialize test db: %v", err)
	}
	defer testDb.CloseDb()
	handler := createFakeServer(testDb)

	// Create two recipes and get their ids
	var ids []string
	for _, name := range []string{"A", "B"} {
		body := DesignRecipeRequest{Name: name, Ingredients: []string{"ing"}, Steps: []string{"s"}}
		req, _ := createRequestWithBody("POST", "/recipes", body)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("setup create: expected 201, got %d", rec.Code)
		}
		var res DesignCreateRecipeResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("decode: %v", err)
		}
		ids = append(ids, res.ID)
	}

	mealBody := DesignCreateMealPlanRequest{Recipes: ids}
	req, _ := createRequestWithBody("POST", "/meal-plans", mealBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Errorf("POST /meal-plans: expected 201, got %d body: %s", rec.Code, getResponseBody(rec))
	}
	var res DesignCreateMealPlanResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if res.ID == "" {
		t.Error("POST /meal-plans: expected non-empty id")
	}
	if res.Message != "Meal plan created successfully" {
		t.Errorf("POST /meal-plans: expected message 'Meal plan created successfully', got %q", res.Message)
	}
}

func TestServer_e2e_design_GET_grocery_list(t *testing.T) {
	testDb, err := initTestDb()
	if err != nil {
		t.Fatalf("unable to initialize test db: %v", err)
	}
	defer testDb.CloseDb()
	handler := createFakeServer(testDb)

	// Create recipes and meal plan
	var ids []string
	body1 := DesignRecipeRequest{
		Name:        "Recipe One",
		Ingredients: []string{"2 eggs", "1 cup milk"},
		Steps:       []string{"Mix", "Cook"},
	}
	req, _ := createRequestWithBody("POST", "/recipes", body1)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create recipe 1: %d", rec.Code)
	}
	var r1 DesignCreateRecipeResponse
	json.NewDecoder(rec.Body).Decode(&r1)
	ids = append(ids, r1.ID)

	body2 := DesignRecipeRequest{
		Name:        "Recipe Two",
		Ingredients: []string{"3 eggs", "2 cups flour"},
		Steps:       []string{"Combine", "Bake"},
	}
	req, _ = createRequestWithBody("POST", "/recipes", body2)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create recipe 2: %d", rec.Code)
	}
	var r2 DesignCreateRecipeResponse
	json.NewDecoder(rec.Body).Decode(&r2)
	ids = append(ids, r2.ID)

	mealBody := DesignCreateMealPlanRequest{Recipes: ids}
	req, _ = createRequestWithBody("POST", "/meal-plans", mealBody)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create meal plan: %d", rec.Code)
	}
	var mealRes DesignCreateMealPlanResponse
	json.NewDecoder(rec.Body).Decode(&mealRes)

	req, _ = http.NewRequest("GET", "/grocery-list/"+mealRes.ID, nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("GET /grocery-list/{id}: expected 200, got %d body: %s", rec.Code, getResponseBody(rec))
	}
	var grocery DesignGroceryListResponse
	if err := json.NewDecoder(rec.Body).Decode(&grocery); err != nil {
		t.Fatalf("decode grocery list: %v", err)
	}
	// Should contain ingredients from both recipes (design: aggregated list)
	if len(grocery.Ingredients) < 2 {
		t.Errorf("GET /grocery-list: expected at least 2 ingredients, got %d: %v", len(grocery.Ingredients), grocery.Ingredients)
	}
}
