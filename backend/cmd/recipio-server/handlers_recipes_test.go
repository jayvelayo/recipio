package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	rec "github.com/jayvelayo/recipio/internal/recipes"
)

func TestCreateRecipeHandler(t *testing.T) {
	handler := createFakeServer(&rec.MockRecipeDatabase{})

	t.Run("Returns 415 without JSON content-type", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/recipes", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusUnsupportedMediaType {
			t.Errorf("Expected status code %d, got %d", http.StatusUnsupportedMediaType, response.Code)
		}
	})

	t.Run("Returns 400 when required fields are missing", func(t *testing.T) {
		body := map[string]interface{}{"description": "A delicious recipe"}
		req, _ := createRequestWithBody("POST", "/recipes", body)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, response.Code)
		}
	})

	t.Run("Creates a new recipe", func(t *testing.T) {
		body := map[string]interface{}{
			"name":         "My Recipe",
			"ingredients":  []map[string]string{{"name": "Banana", "quantity": "2"}, {"name": "Egg", "quantity": "2"}},
			"instructions": []string{"Cook", "Serve"},
		}
		req, _ := createRequestWithBody("POST", "/recipes", body)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d", http.StatusCreated, response.Code)
		}
	})
}

func TestGetRecipeHandler(t *testing.T) {
	mockdb := rec.MockRecipeDatabase{}
	mockdb.CreateRecipe(rec.Recipe{Name: "Test Recipe 1", Ingredients: []rec.Ingredient{{Name: "Salt", Quantity: "1 tsp"}}, Instructions: rec.InstructionList{"Mix"}})
	mockdb.CreateRecipe(rec.Recipe{Name: "Test Recipe 2", Ingredients: []rec.Ingredient{{Name: "Pepper", Quantity: "1 tsp"}}, Instructions: rec.InstructionList{"Mix"}})
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

	t.Run("Returns 404 for non-existent recipe", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/recipes/999", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, response.Code)
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

	t.Run("Returns 415 without JSON content-type", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/recipes/1", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusUnsupportedMediaType {
			t.Errorf("Expected status code %d, got %d", http.StatusUnsupportedMediaType, response.Code)
		}
	})

	t.Run("Returns 400 when required fields are missing", func(t *testing.T) {
		body := map[string]interface{}{"name": "No Ingredients Recipe"}
		req, _ := createRequestWithBody("PUT", "/recipes/1", body)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, response.Code)
		}
	})

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
}

func TestDeleteRecipeHandler(t *testing.T) {
	mockdb := rec.MockRecipeDatabase{}
	mockdb.CreateRecipe(rec.Recipe{
		Name:         "Recipe to Delete",
		Ingredients:  []rec.Ingredient{{Name: "Salt", Quantity: "1 tsp"}},
		Instructions: rec.InstructionList{"Sprinkle"},
	})
	handler := createFakeServer(&mockdb)

	t.Run("Returns 400 for invalid recipe ID", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/recipes/abc", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, response.Code)
		}
	})

	t.Run("Returns 404 for non-existent recipe", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/recipes/999", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, response.Code)
		}
	})

	t.Run("Deletes an existing recipe", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/recipes/1", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusNoContent {
			t.Errorf("Expected status code %d, got %d: %s", http.StatusNoContent, response.Code, getResponseBody(response))
		}
	})
}
