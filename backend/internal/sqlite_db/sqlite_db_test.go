package sqlite_db_test

import (
	"testing"

	rec "github.com/jayvelayo/recipio/internal/recipes"
)

func TestCreateRecipe(t *testing.T) {
	t.Run("Recipe name cannot be empty", func(t *testing.T) {
		recipe := rec.Recipe{
			Description: "A delicious recipe",
		}
		db := &rec.MockRecipeDatabase{}
		id, err := db.CreateRecipe(recipe)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
		if id != 0 {
			t.Errorf("Expected 0, got: %d", id)
		}
	})

	// Is this needed if the check is within the mock function, not the actual function?
	t.Run("Recipe with this name already exists", func(t *testing.T) {
		recipe := rec.Recipe{
			Name:        "My Recipe",
			Description: "A delicious recipe",
		}
		db := &rec.MockRecipeDatabase{}
		db.CreateRecipe(recipe)
		id, err := db.CreateRecipe(recipe)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
		if id != 0 {
			t.Errorf("Expected 0, got: %d", id)
		}
	})

	t.Run("Creates a new recipe", func(t *testing.T) {
		recipe := rec.Recipe{
			Name:        "My Recipe",
			Description: "A delicious recipe",
		}
		db := &rec.MockRecipeDatabase{}
		id, err := db.CreateRecipe(recipe)
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
		db := &rec.MockRecipeDatabase{}
		recipe := rec.Recipe{
			Name:        "My Recipe",
			Description: "A delicious recipe",
		}
		id, err := db.CreateRecipe(recipe)
		if err != nil {
			t.Fatalf("Failed to create recipe: %v", err)
		}
		if err := db.DeleteRecipe(int(id)); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if db.GetRecipeCount() != 0 {
			t.Errorf("Expected recipeCount 0, got %d", db.GetRecipeCount())
		}
	})

	t.Run("Delete non-existent recipe", func(t *testing.T) {
		db := &rec.MockRecipeDatabase{}
		err := db.DeleteRecipe(1)
		if err == nil {
			t.Errorf("Expected error for non-existent recipe, got nil")
		}
	})
}
