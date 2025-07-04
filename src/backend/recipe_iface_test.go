package main

import (
	"fmt"
	"testing"
)

type MockRecipeDatabase struct {
	recipes     []Recipe
	recipeCount int
}

func (db *MockRecipeDatabase) createRecipe(newRecipe Recipe) (uint64, error) {
	if newRecipe.Name == "" {
		return 0, fmt.Errorf("recipe name cannot be empty")
	}
	for _, recipe := range db.recipes {
		if newRecipe.Name == recipe.Name {
			return 0, fmt.Errorf("recipe with this name already exists")
		}
	}
	newRecipe.ID = db.recipeCount + 1
	db.recipeCount++
	db.recipes = append(db.recipes, newRecipe)
	return uint64(newRecipe.ID), nil
}

func (db *MockRecipeDatabase) getRecipe(id int) (Recipe, error) {
	if id < 1 || id > db.recipeCount {
		return Recipe{}, fmt.Errorf("recipe not found")
	}
	return db.recipes[id-1], nil
}

func (db *MockRecipeDatabase) getAllRecipes() (Recipes, error) {
	return db.recipes, nil
}

func (db *MockRecipeDatabase) deleteRecipe(id int) error {
	if id < 1 || id > db.recipeCount {
		return fmt.Errorf("recipe not found")
	}
	// Remove the recipe from the slice
	db.recipes = append(db.recipes[:id-1], db.recipes[id:]...)
	db.recipeCount--
	return nil
}

func (db *MockRecipeDatabase) closeDb() {
}

func TestCreateRecipe(t *testing.T) {
	t.Run("Recipe name cannot be empty", func(t *testing.T) {
		recipe := Recipe{
			Description: "A delicious recipe",
		}
		db := &MockRecipeDatabase{}
		id, err := db.createRecipe(recipe)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
		if id != 0 {
			t.Errorf("Expected 0, got: %d", id)
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
		id, err := db.createRecipe(recipe)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
		if id != 0 {
			t.Errorf("Expected 0, got: %d", id)
		}
	})

	t.Run("Creates a new recipe", func(t *testing.T) {
		recipe := Recipe{
			Name:        "My Recipe",
			Description: "A delicious recipe",
		}
		db := &MockRecipeDatabase{}
		id, err := db.createRecipe(recipe)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if id != 1 {
			t.Errorf("Expected 1, got: %d", id)
		}
	})
}

func TestDeleteRecipe(t *testing.T) {
	t.Run("Delete existing recipe", func(t *testing.T) {
		db := &MockRecipeDatabase{}
		recipe := Recipe{
			Name:        "My Recipe",
			Description: "A delicious recipe",
		}
		id, err := db.createRecipe(recipe)
		if err != nil {
			t.Fatalf("Failed to create recipe: %v", err)
		}
		if err := db.deleteRecipe(int(id)); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if db.recipeCount != 0 {
			t.Errorf("Expected recipeCount 0, got %d", db.recipeCount)
		}
	})

	t.Run("Delete non-existent recipe", func(t *testing.T) {
		db := &MockRecipeDatabase{}
		err := db.deleteRecipe(1)
		if err == nil {
			t.Errorf("Expected error for non-existent recipe, got nil")
		}
	})
}
