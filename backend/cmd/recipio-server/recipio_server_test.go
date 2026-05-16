package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	rec "github.com/jayvelayo/recipio/internal/recipes"
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
	// Use empty allowed origins for testing - CORS will be bypassed in tests
	allowedOrigins := []string{}
	mux := http.NewServeMux()
	SetUpRoutes(mux, db, allowedOrigins)
	return mux
}

func TestCreateRecipeHandler(t *testing.T) {
	handler := createFakeServer(&rec.MockRecipeDatabase{})
	t.Run("Testing that endpoint must accept json", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/recipes", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusUnsupportedMediaType {
			t.Errorf("Expected status code %d, got %d", http.StatusUnsupportedMediaType, response.Code)
		}
	})

	t.Run("Testing that json body has required fields", func(t *testing.T) {
		incorrectBody := map[string]interface{}{
			"description": "A delicious recipe",
		}
		req, _ := createRequestWithBody("POST", "/recipes", incorrectBody)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, response.Code)
		}
	})

	t.Run("Creates a new recipe", func(t *testing.T) {
		correctBody := map[string]interface{}{
			"name": "My Recipe",
			"ingredients": []map[string]string{
				{"name": "Banana", "quantity": "2"},
				{"name": "Egg", "quantity": "2"},
			},
			"instructions": []string{"Cook", "me"},
		}
		req, _ := createRequestWithBody("POST", "/recipes", correctBody)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, req)
		if response.Code != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d", http.StatusCreated, response.Code)
		}
	})
}

func TestUpdateRecipeHandler(t *testing.T) {
	mockdb := rec.MockRecipeDatabase{}
	mockdb.CreateRecipe(rec.Recipe{
		Name:         "Original Recipe",
		Ingredients:  []rec.Ingredient{{Name: "Flour", Quantity: "2 cups"}},
		Instructions: rec.InstructionList{"Mix", "Bake"},
	})
	handler := createFakeServer(&mockdb)

	t.Run("Updates an existing recipe", func(t *testing.T) {
		body := map[string]interface{}{
			"name":         "Updated Recipe",
			"ingredients":  []map[string]string{{"name": "Sugar", "quantity": "1 cup"}},
			"instructions": []string{"Mix well", "Bake longer"},
		}
		req, _ := createRequestWithBody("PUT", "/recipes/1", body)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusNoContent {
			t.Errorf("Expected status code %d, got %d: %s", http.StatusNoContent, response.Code, getResponseBody(response))
		}
	})

	t.Run("Returns 404 for non-existent recipe", func(t *testing.T) {
		body := map[string]interface{}{
			"name":         "Updated Recipe",
			"ingredients":  []map[string]string{{"name": "Sugar", "quantity": "1 cup"}},
			"instructions": []string{"Mix well"},
		}
		req, _ := createRequestWithBody("PUT", "/recipes/999", body)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, response.Code)
		}
	})

	t.Run("Returns 415 without JSON content-type", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/recipes/1", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusUnsupportedMediaType {
			t.Errorf("Expected status code %d, got %d", http.StatusUnsupportedMediaType, response.Code)
		}
	})

	t.Run("Returns 400 when required fields are missing", func(t *testing.T) {
		body := map[string]interface{}{
			"name": "No Ingredients Recipe",
		}
		req, _ := createRequestWithBody("PUT", "/recipes/1", body)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, response.Code)
		}
	})
}

func TestGetRecipeHandler(t *testing.T) {
	mockdb := rec.MockRecipeDatabase{}
	// Add some mock recipes
	recipe1 := rec.Recipe{ID: 1, Name: "Test Recipe 1"}
	recipe2 := rec.Recipe{ID: 2, Name: "Test Recipe 2"}
	mockdb.CreateRecipe(recipe1)
	mockdb.CreateRecipe(recipe2)

	handler := createFakeServer(&mockdb)
	t.Run("Fetches a recipe by ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/recipes/1", nil)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, req)
		if response.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, response.Code)
		}
	})

	t.Run("Fetches all recipes", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/recipes", nil)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, req)
		if response.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, response.Code)
		}
	})

	t.Run("Checks if id is not found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/recipes/999", nil)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, req)
		if response.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, response.Code)
		}
	})
}
