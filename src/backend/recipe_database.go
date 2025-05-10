package main

import (
	"fmt"
)

type LocalRecipeDatabase struct {
	recipes     []Recipe
	recipeCount int
}

func (db *LocalRecipeDatabase) createRecipe(newRecipe Recipe) error {
	if newRecipe.Name == "" {
		return fmt.Errorf("recipe name cannot be empty")
	}
	for _, recipe := range db.recipes {
		if newRecipe.Name == recipe.Name {
			return fmt.Errorf("recipe with this name already exists")
		}
	}
	newRecipe.ID = db.recipeCount + 1
	db.recipeCount++
	db.recipes = append(db.recipes, newRecipe)
	return nil
}

func (db *LocalRecipeDatabase) getRecipe(id int) (Recipe, error) {
	if id < 1 || id > db.recipeCount {
		return Recipe{}, fmt.Errorf("recipe not found")
	}
	return db.recipes[id-1], nil
}
