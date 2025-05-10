package main

import (
	"fmt"
)

type Recipe struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Ingredients  []string `json:"ingredients"`
	Instructions string   `json:"instructions"`
}

type RecipeDatabase interface {
	createRecipe(recipe Recipe) error
	getRecipe(id int) (Recipe, error)
}

func CreateRecipe(recipe Recipe, db RecipeDatabase) error {
	if recipe.Name == "" {
		return fmt.Errorf("recipe name cannot be empty")
	}
	return db.createRecipe(recipe)
}

func GetRecipe(id int, db RecipeDatabase) (Recipe, error) {
	if id < 1 {
		return Recipe{}, fmt.Errorf("id is not valid")
	}
	return db.getRecipe(id)
}
