package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	rec "github.com/jayvelayo/recipio/internal/recipes"
)

func TestGetGroceryListForMealPlanHandler(t *testing.T) {
	mockdb := rec.MockRecipeDatabase{}
	mockdb.CreateMealPlan(testUserID, []string{"1", "2"})
	handler := createFakeServer(&mockdb)

	t.Run("Returns ingredients for an existing meal plan", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/grocery-list/1", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d: %s", http.StatusOK, response.Code, getResponseBody(response))
		}
	})

	t.Run("Returns 404 for non-existent meal plan", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/grocery-list/999", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, response.Code)
		}
	})
}

func TestCreateGroceryListHandler(t *testing.T) {
	handler := createFakeServer(&rec.MockRecipeDatabase{})

	t.Run("Returns 415 without JSON content-type", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/grocery-lists", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusUnsupportedMediaType {
			t.Errorf("Expected status code %d, got %d", http.StatusUnsupportedMediaType, response.Code)
		}
	})

	t.Run("Returns 400 when name is missing", func(t *testing.T) {
		body := map[string]interface{}{"items": []map[string]interface{}{}}
		req, _ := createRequestWithBody("POST", "/grocery-lists", body)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, response.Code)
		}
	})

	t.Run("Creates a new grocery list", func(t *testing.T) {
		body := map[string]interface{}{
			"name":  "Weekly Groceries",
			"items": []map[string]interface{}{{"name": "Eggs", "quantity": "12", "checked": false}},
		}
		req, _ := createRequestWithBody("POST", "/grocery-lists", body)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d: %s", http.StatusCreated, response.Code, getResponseBody(response))
		}
	})
}

func TestGetAllGroceryListsHandler(t *testing.T) {
	mockdb := rec.MockRecipeDatabase{}
	mockdb.CreateGroceryList(testUserID, "List 1", []rec.GroceryListItem{{Name: "Eggs", Quantity: "12"}}, nil)
	handler := createFakeServer(&mockdb)

	t.Run("Returns 200 with all grocery lists", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/grocery-lists", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, response.Code)
		}
	})
}

func TestGetGroceryListByIDHandler(t *testing.T) {
	mockdb := rec.MockRecipeDatabase{}
	mockdb.CreateGroceryList(testUserID, "My List", []rec.GroceryListItem{{Name: "Milk", Quantity: "1 gallon"}}, nil)
	handler := createFakeServer(&mockdb)

	t.Run("Returns 200 for existing grocery list", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/grocery-lists/1", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d: %s", http.StatusOK, response.Code, getResponseBody(response))
		}
	})

	t.Run("Returns 404 for non-existent grocery list", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/grocery-lists/999", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, response.Code)
		}
	})
}

func TestUpdateGroceryListHandler(t *testing.T) {
	mockdb := rec.MockRecipeDatabase{}
	mockdb.CreateGroceryList(testUserID, "List to Update", []rec.GroceryListItem{{Name: "Bread", Quantity: "1 loaf"}}, nil)
	handler := createFakeServer(&mockdb)

	t.Run("Returns 415 without JSON content-type", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/grocery-lists/1", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusUnsupportedMediaType {
			t.Errorf("Expected status code %d, got %d", http.StatusUnsupportedMediaType, response.Code)
		}
	})

	t.Run("Updates an existing grocery list", func(t *testing.T) {
		items := []map[string]interface{}{{"name": "Bread", "quantity": "2 loaves", "checked": true}}
		req, _ := createRequestWithBody("PUT", "/grocery-lists/1", items)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusNoContent {
			t.Errorf("Expected status code %d, got %d: %s", http.StatusNoContent, response.Code, getResponseBody(response))
		}
	})

	t.Run("Returns 404 for non-existent grocery list", func(t *testing.T) {
		items := []map[string]interface{}{{"name": "Bread", "quantity": "1 loaf", "checked": false}}
		req, _ := createRequestWithBody("PUT", "/grocery-lists/999", items)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, response.Code)
		}
	})
}

func TestDeleteGroceryListHandler(t *testing.T) {
	mockdb := rec.MockRecipeDatabase{}
	mockdb.CreateGroceryList(testUserID, "List to Delete", []rec.GroceryListItem{{Name: "Cheese", Quantity: "500g"}}, nil)
	handler := createFakeServer(&mockdb)

	t.Run("Returns 404 for non-existent grocery list", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/grocery-lists/999", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, response.Code)
		}
	})

	t.Run("Deletes an existing grocery list", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/grocery-lists/1", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusNoContent {
			t.Errorf("Expected status code %d, got %d: %s", http.StatusNoContent, response.Code, getResponseBody(response))
		}
	})
}
