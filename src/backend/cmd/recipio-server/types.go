package main

import (
	rec "github.com/jayvelayo/recipio/internal/recipes"
)

// Design API types (doc/server_design.md)
type designRecipeRequest struct {
	Name        string   `json:"name"`
	Ingredients []string `json:"ingredients"`
	Steps       []string `json:"steps"`
}

type designCreateRecipeResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type designRecipeResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Ingredients []string `json:"ingredients"`
	Steps       []string `json:"steps"`
}

type designCreateMealPlanRequest struct {
	Recipes []string `json:"recipes"`
}

type designCreateMealPlanResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type designGroceryListResponse struct {
	Ingredients []string `json:"ingredients"`
}

type RecipeBody struct {
	Name         string   `json:"name"`
	Ingredients  []string `json:"ingredients"`
	Instructions []string `json:"instructions"`
}

type CreateRecipeResponse struct {
	Status       ResponseStatus `json:"status"`
	ErrorMessage string         `json:"errorMessage"`
	RecipeId     uint64         `json:"id"`
}

type ResponseStatus int

const (
	StatusOK ResponseStatus = iota
	StatusError
	StatusEmptyBody
	StatusInvalidJson
	StatusNotFound
	StatusMissingFields
	StatusAlreadyExist
	StatusEncodingError
)

type designCreateGroceryListRequest struct {
	Name       string                `json:"name"`
	Items      []rec.GroceryListItem `json:"items"`
	MealPlanID *string               `json:"meal_plan_id,omitempty"`
}

type designCreateGroceryListResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
