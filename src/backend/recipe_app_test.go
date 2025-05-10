package main

import (
	"testing"
	"fmt"
)

type MockRecipeDatabase struct {
	recipes []Recipe
	recipeCount int
}

func (db *MockRecipeDatabase) createRecipe(newRecipe Recipe) error {
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

func (db *MockRecipeDatabase) getRecipe(id int) (Recipe, error) {
	if id < 1 || id > db.recipeCount {
		return Recipe{}, fmt.Errorf("recipe not found")
	}
	return db.recipes[id-1], nil
}

func TestCreateRecipe(t *testing.T) {
	t.Run("Recipe name cannot be empty", func(t *testing.T) {
		recipe := Recipe{
			Description: "A delicious recipe",
		}
		db := &MockRecipeDatabase{}
		err := CreateRecipe(recipe, db)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})

	// Is this needed if the check is within the mock function, not the actual function?
	t.Run("Recipe with this name already exists", func(t *testing.T) {
		recipe := Recipe{
			Name:        "My Recipe",
			Description: "A delicious recipe",
		}
		db := &MockRecipeDatabase{}
		db.createRecipe(recipe)
		err := CreateRecipe(recipe, db)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})

	t.Run("Creates a new recipe", func(t *testing.T) {
		recipe := Recipe{
			Name:        "My Recipe",
			Description: "A delicious recipe",
		}
		db := &MockRecipeDatabase{}
		err := CreateRecipe(recipe, db)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}