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

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	Email string `json:"email"`
}

type UserInfoResponse struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
