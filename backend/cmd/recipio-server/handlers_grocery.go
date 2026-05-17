package main

import (
	"encoding/json"
	"net/http"

	rec "github.com/jayvelayo/recipio/internal/recipes"
)

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
		encodeJson(w, http.StatusOK, GroceryListResponse{Ingredients: ingredients})
	})
}

func handleDesignCreateGroceryList(recipeDb rec.RecipeDatabase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}
		var body CreateGroceryListRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}
		if body.Name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}
		userID := r.Context().Value(userIDKey).(string)
		id, err := recipeDb.CreateGroceryList(userID, body.Name, body.Items, body.MealPlanID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		encodeJson(w, http.StatusCreated, CreateGroceryListResponse{ID: id, Name: body.Name})
	})
}

func handleDesignGetAllGroceryLists(recipeDb rec.RecipeDatabase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(userIDKey).(string)
		lists, err := recipeDb.GetAllGroceryLists(userID)
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
