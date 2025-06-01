package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-test/deep"
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

func createFakeServer(db RecipeDatabase) http.Handler {
	mux := http.NewServeMux()
	SetUpRoutes(mux, db)
	return mux
}

func TestCreateRecipeHandler(t *testing.T) {
	handler := createFakeServer(&MockRecipeDatabase{})
	t.Run("Testing that endpoint must accept json", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/v1/recipe", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusUnsupportedMediaType {
			t.Errorf("Expected status code %d, got %d", http.StatusUnsupportedMediaType, response.Code)
		}
	})

	t.Run("Testing that json body has the name keys", func(t *testing.T) {
		incorrectBody := Recipe{
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
		correctBody := Recipe{
			Name:        "My Recipe",
			Description: "A delicious recipe",
			Ingredients: []Ingredient{
				{Name: "Banana", Quantity: "2"},
				{Name: "Egg", Quantity: "2"},
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

	mockdb := MockRecipeDatabase{}
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

var mockRecipes = Recipes{
	{
		ID:   1,
		Name: "French toast",
		Ingredients: []Ingredient{
			{Name: "eggs", Quantity: "2 pieces"},
			{Name: "bread", Quantity: "2 pieces"},
			{Name: "milk", Quantity: "50 mL"},
		},
		Instructions: instructionList{
			"Beats eggs until smooth",
			"Add milk to the egg mixture",
			"Dip the bread into the milk-egg mixture",
			"Cook on a non-stick pan with butter until desired texture",
		},
	},
	{
		ID:   2,
		Name: "Pancakes",
		Ingredients: []Ingredient{
			{Name: "flour", Quantity: "200 grams"},
			{Name: "milk", Quantity: "300 mL"},
			{Name: "egg", Quantity: "1 pieces"},
			{Name: "sugar", Quantity: "2 tablespoons"},
			{Name: "baking powder", Quantity: "1 teaspoons"},
		},
		Instructions: instructionList{
			"Mix all dry ingredients in a bowl",
			"Add milk and egg, then whisk until smooth",
			"Heat a lightly oiled pan over medium heat",
			"Pour batter onto the pan and cook until bubbles form, then flip and cook until golden",
		},
	},
	{
		ID:   3,
		Name: "Scrambled Eggs",
		Ingredients: []Ingredient{
			{Name: "eggs", Quantity: "3 pieces"},
			{Name: "milk", Quantity: "30 mL"},
			{Name: "butter", Quantity: "1 tablespoons"},
			{Name: "salt", Quantity: "0.5 teaspoons"},
		},
		Instructions: instructionList{
			"Crack the eggs into a bowl and add milk and salt",
			"Whisk the mixture until well combined",
			"Melt butter in a pan over medium heat",
			"Pour in the egg mixture and stir gently until just set",
		},
	},
	{
		ID:   4,
		Name: "Grilled Cheese Sandwich",
		Ingredients: []Ingredient{
			{Name: "bread", Quantity: "2 pieces"},
			{Name: "cheese slices", Quantity: "2 pieces"},
			{Name: "butter", Quantity: "1 tablespoons"},
		},
		Instructions: instructionList{
			"Butter one side of each bread slice",
			"Place cheese between the unbuttered sides of the bread",
			"Cook in a pan over medium heat until golden on both sides and cheese is melted",
		},
	},
	{
		ID:   5,
		Name: "Banana Smoothie",
		Ingredients: []Ingredient{
			{Name: "banana", Quantity: "1 pieces"},
			{Name: "milk", Quantity: "200 mL"},
			{Name: "yogurt", Quantity: "100 grams"},
			{Name: "honey", Quantity: "1 tablespoons"},
		},
		Instructions: instructionList{
			"Peel and slice the banana",
			"Add banana, milk, yogurt, and honey to a blender",
			"Blend until smooth",
			"Serve chilled",
		},
	},
}

func initTestDb() (RecipeDatabase, error) {
	db, err := sql.Open("sqlite", "recipes.db")
	if err != nil {
		log.Fatal(err)
	}
	schemaData := SchemaData{
		RecipesTable:     "test_recipes",
		IngredientsTable: "test_ingredients",
	}
	db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", schemaData.RecipesTable))
	db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", schemaData.IngredientsTable))
	schema, err := applySchema("./schema.tmpl", schemaData)
	if err != nil {
		return nil, fmt.Errorf("unable to create schema db: %v", err)
	}
	_, err = db.Exec(schema)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize db: %v", err)
	}
	sqliteDb := &SqliteDatabaseContext{
		sqliteDb: db,
		schema:   schemaData,
	}
	return sqliteDb, nil
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
	defer testDb.closeDb()
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
		recipes, err := decodeJsonResponse[Recipes](response.Body)
		if err != nil {
			t.Fatalf("failed to decode response body: %s", err)
		}
		if len(recipes) != 1 {
			t.Errorf("Expected length of recipes arr %d, got %d", 1, len(recipes))
		}
		diff := deep.Equal(recipes[0], mockRecipes[1])
		if diff != nil {
			t.Error(diff)
		}
	})
}

func TestServer_e2e_multiple_recipes(t *testing.T) {
	testDb, err := initTestDb()
	if err != nil {
		t.Fatalf("unable to initialize test db %v", err)
	}
	defer testDb.closeDb()
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
		recipes, err := decodeJsonResponse[Recipes](response.Body)
		if err != nil {
			t.Fatalf("failed to decode response body: %s", err)
		}
		if len(recipes) != 5 {
			t.Errorf("Expected length of recipes arr %d, got %d", 5, len(recipes))
		}
		diff := deep.Equal(recipes, mockRecipes)
		if diff != nil {
			t.Error(diff)
		}
	})
}
