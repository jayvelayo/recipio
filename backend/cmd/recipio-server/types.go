package main

import (
	rec "github.com/jayvelayo/recipio/internal/recipes"
)

type CreateRecipeResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type CreateMealPlanRequest struct {
	Recipes []string `json:"recipes"`
}

type CreateMealPlanResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type GroceryListResponse struct {
	Ingredients []string `json:"ingredients"`
}

type CreateGroceryListRequest struct {
	Name       string                `json:"name"`
	Items      []rec.GroceryListItem `json:"items"`
	MealPlanID *string               `json:"meal_plan_id,omitempty"`
}

type CreateGroceryListResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
