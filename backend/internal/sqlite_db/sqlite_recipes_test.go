package sqlite_db_test

import (
	"testing"

	rec "github.com/jayvelayo/recipio/internal/recipes"
	"github.com/jayvelayo/recipio/internal/sqlite_db"
)

func initRecipeDB(t *testing.T) rec.RecipeDatabase {
	t.Helper()
	db, err := sqlite_db.InitDb(":memory:")
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	t.Cleanup(db.CloseDb)
	return db
}

func TestCreateRecipe(t *testing.T) {
	t.Run("creates a recipe and returns a non-zero ID", func(t *testing.T) {
		db := initRecipeDB(t)
		recipe := rec.Recipe{
			Name:         "Pasta",
			Instructions: rec.InstructionList{"Boil water", "Cook pasta"},
			Ingredients:  []rec.Ingredient{{Name: "pasta", Quantity: "200g"}},
		}
		id, err := db.CreateRecipe("test-user", recipe)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id == 0 {
			t.Error("expected non-zero ID")
		}
	})

	t.Run("IDs increment for successive recipes", func(t *testing.T) {
		db := initRecipeDB(t)
		id1, _ := db.CreateRecipe("test-user", rec.Recipe{Name: "Recipe A", Instructions: rec.InstructionList{"Step 1"}})
		id2, _ := db.CreateRecipe("test-user", rec.Recipe{Name: "Recipe B", Instructions: rec.InstructionList{"Step 1"}})
		if id2 <= id1 {
			t.Errorf("expected id2 (%d) > id1 (%d)", id2, id1)
		}
	})

	t.Run("returns error when an ingredient has an empty name", func(t *testing.T) {
		db := initRecipeDB(t)
		recipe := rec.Recipe{
			Name:         "Omelette",
			Instructions: rec.InstructionList{"Crack eggs"},
			Ingredients:  []rec.Ingredient{{Name: "", Quantity: "2"}},
		}
		_, err := db.CreateRecipe("test-user", recipe)
		if err == nil {
			t.Error("expected error for empty ingredient name, got nil")
		}
	})
}

func TestGetRecipe(t *testing.T) {
	t.Run("retrieves the created recipe with ingredients and instructions", func(t *testing.T) {
		db := initRecipeDB(t)
		recipe := rec.Recipe{
			Name:         "Scrambled Eggs",
			Instructions: rec.InstructionList{"Crack eggs", "Whisk", "Cook"},
			Ingredients: []rec.Ingredient{
				{Name: "eggs", Quantity: "3"},
				{Name: "butter", Quantity: "1 tbsp"},
			},
		}
		id, err := db.CreateRecipe("test-user", recipe)
		if err != nil {
			t.Fatalf("failed to create recipe: %v", err)
		}

		got, err := db.GetRecipe(int(id))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name != recipe.Name {
			t.Errorf("expected name %q, got %q", recipe.Name, got.Name)
		}
		if len(got.Instructions) != len(recipe.Instructions) {
			t.Errorf("expected %d instructions, got %d", len(recipe.Instructions), len(got.Instructions))
		}
		if len(got.Ingredients) != len(recipe.Ingredients) {
			t.Errorf("expected %d ingredients, got %d", len(recipe.Ingredients), len(got.Ingredients))
		}
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		db := initRecipeDB(t)
		_, err := db.GetRecipe(9999)
		if err == nil {
			t.Error("expected error for non-existent recipe, got nil")
		}
	})
}

func TestGetAllRecipes(t *testing.T) {
	t.Run("returns empty slice when no recipes exist", func(t *testing.T) {
		db := initRecipeDB(t)
		recipes, err := db.GetAllRecipes("test-user")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(recipes) != 0 {
			t.Errorf("expected 0 recipes, got %d", len(recipes))
		}
	})

	t.Run("returns all created recipes", func(t *testing.T) {
		db := initRecipeDB(t)
		db.CreateRecipe("test-user", rec.Recipe{Name: "Recipe 1", Instructions: rec.InstructionList{"Step 1"}})
		db.CreateRecipe("test-user", rec.Recipe{Name: "Recipe 2", Instructions: rec.InstructionList{"Step 1"}})
		db.CreateRecipe("test-user", rec.Recipe{Name: "Recipe 3", Instructions: rec.InstructionList{"Step 1"}})

		recipes, err := db.GetAllRecipes("test-user")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(recipes) != 3 {
			t.Errorf("expected 3 recipes, got %d", len(recipes))
		}
	})
}

func TestUpdateRecipe(t *testing.T) {
	t.Run("updates name, instructions, and ingredients", func(t *testing.T) {
		db := initRecipeDB(t)
		id, err := db.CreateRecipe("test-user", rec.Recipe{
			Name:         "Old Name",
			Instructions: rec.InstructionList{"Old step"},
			Ingredients:  []rec.Ingredient{{Name: "old ingredient", Quantity: "1"}},
		})
		if err != nil {
			t.Fatalf("failed to create recipe: %v", err)
		}

		updated := rec.Recipe{
			Name:         "New Name",
			Instructions: rec.InstructionList{"New step 1", "New step 2"},
			Ingredients:  []rec.Ingredient{{Name: "new ingredient", Quantity: "2 cups"}},
		}
		if err := db.UpdateRecipe(int(id), updated); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, err := db.GetRecipe(int(id))
		if err != nil {
			t.Fatalf("unexpected error fetching updated recipe: %v", err)
		}
		if got.Name != updated.Name {
			t.Errorf("expected name %q, got %q", updated.Name, got.Name)
		}
		if len(got.Instructions) != len(updated.Instructions) {
			t.Errorf("expected %d instructions, got %d", len(updated.Instructions), len(got.Instructions))
		}
		if len(got.Ingredients) != 1 || got.Ingredients[0].Name != "new ingredient" {
			t.Errorf("ingredients not updated correctly: %+v", got.Ingredients)
		}
	})

	t.Run("returns error for non-existent recipe ID", func(t *testing.T) {
		db := initRecipeDB(t)
		err := db.UpdateRecipe(9999, rec.Recipe{Name: "Ghost", Instructions: rec.InstructionList{"Step"}})
		if err == nil {
			t.Error("expected error for non-existent recipe, got nil")
		}
	})
}

func TestDeleteRecipe(t *testing.T) {
	t.Run("deletes an existing recipe", func(t *testing.T) {
		db := initRecipeDB(t)
		id, err := db.CreateRecipe("test-user", rec.Recipe{
			Name:         "To Delete",
			Instructions: rec.InstructionList{"Step 1"},
			Ingredients:  []rec.Ingredient{{Name: "sugar", Quantity: "1 cup"}},
		})
		if err != nil {
			t.Fatalf("failed to create recipe: %v", err)
		}

		if err := db.DeleteRecipe(int(id)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = db.GetRecipe(int(id))
		if err == nil {
			t.Error("expected error fetching deleted recipe, got nil")
		}
	})

	t.Run("returns error for non-existent recipe", func(t *testing.T) {
		db := initRecipeDB(t)
		err := db.DeleteRecipe(9999)
		if err == nil {
			t.Error("expected error for non-existent recipe, got nil")
		}
	})
}
