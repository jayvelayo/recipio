package main

import (
	"encoding/json"
	"net/http"

	rec "github.com/jayvelayo/recipio/internal/recipes"
)

func handleDesignGetAllMealPlans(recipeDb rec.RecipeDatabase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		plans, err := recipeDb.GetAllMealPlans()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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
