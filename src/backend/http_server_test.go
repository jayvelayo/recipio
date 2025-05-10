package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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

func TestCreateRecipeHandler(t *testing.T) {
	handler := handleCreateRecipe(&MockRecipeDatabase{})

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
	mockdb.recipes = append(mockdb.recipes, Recipe{ID: 1})
	mockdb.recipeCount = 1
	handler := handleGetRecipe(&mockdb)
	t.Run("Fetches a recipe by ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/recipe", nil)
		
		q := req.URL.Query()
		q.Add("id", "1")
		req.URL.RawQuery = q.Encode()
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, req)
		if response.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, response.Code)
		}
	})
	t.Run("Checks if id is found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/recipe", nil)
		
		q := req.URL.Query()
		q.Add("id", "2")
		req.URL.RawQuery = q.Encode()
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, req)
		if response.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, response.Code)
		}
	})
}
