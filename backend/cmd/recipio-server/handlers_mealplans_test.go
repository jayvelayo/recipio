package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	rec "github.com/jayvelayo/recipio/internal/recipes"
)

func TestGetAllMealPlansHandler(t *testing.T) {
	handler := createFakeServer(&rec.MockRecipeDatabase{})

	t.Run("Returns 200 with empty list", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/meal-plans", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, response.Code)
		}
	})
}

func TestCreateMealPlanHandler(t *testing.T) {
	handler := createFakeServer(&rec.MockRecipeDatabase{})

	t.Run("Returns 415 without JSON content-type", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/meal-plans", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusUnsupportedMediaType {
			t.Errorf("Expected status code %d, got %d", http.StatusUnsupportedMediaType, response.Code)
		}
	})

	t.Run("Creates a meal plan with recipe IDs", func(t *testing.T) {
		body := map[string]interface{}{"recipes": []string{"1", "2"}}
		req, _ := createRequestWithBody("POST", "/meal-plans", body)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d: %s", http.StatusCreated, response.Code, getResponseBody(response))
		}
	})
}

func TestDeleteMealPlanHandler(t *testing.T) {
	mockdb := rec.MockRecipeDatabase{}
	mockdb.CreateMealPlan(testUserID, []string{"1", "2"})
	handler := createFakeServer(&mockdb)

	t.Run("Returns 404 for non-existent meal plan", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/meal-plans/999", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, response.Code)
		}
	})

	t.Run("Deletes an existing meal plan", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/meal-plans/1", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusNoContent {
			t.Errorf("Expected status code %d, got %d: %s", http.StatusNoContent, response.Code, getResponseBody(response))
		}
	})
}
