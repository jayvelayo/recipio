package sqlite_db_test

import (
	"testing"

	rec "github.com/jayvelayo/recipio/internal/recipes"
	"github.com/jayvelayo/recipio/internal/sqlite_db"
)

const (
	userA = "user-a"
	userB = "user-b"
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
	t.Run("creates a recipe and returns a non-empty ID", func(t *testing.T) {
		db := initRecipeDB(t)
		recipe := rec.Recipe{
			Name:         "Pasta",
			Instructions: rec.InstructionList{"Boil water", "Cook pasta"},
			Ingredients:  []rec.Ingredient{{Name: "pasta", Quantity: "200g"}},
		}
		id, err := db.CreateRecipe(userA, recipe)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id == "" {
			t.Error("expected non-empty ID")
		}
	})

	t.Run("successive recipes get unique IDs", func(t *testing.T) {
		db := initRecipeDB(t)
		id1, _ := db.CreateRecipe(userA, rec.Recipe{Name: "Recipe A", Instructions: rec.InstructionList{"Step 1"}})
		id2, _ := db.CreateRecipe(userA, rec.Recipe{Name: "Recipe B", Instructions: rec.InstructionList{"Step 1"}})
		if id1 == id2 {
			t.Errorf("expected unique IDs, got %q for both", id1)
		}
	})

	t.Run("returns error when an ingredient has an empty name", func(t *testing.T) {
		db := initRecipeDB(t)
		recipe := rec.Recipe{
			Name:         "Omelette",
			Instructions: rec.InstructionList{"Crack eggs"},
			Ingredients:  []rec.Ingredient{{Name: "", Quantity: "2"}},
		}
		_, err := db.CreateRecipe(userA, recipe)
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
		id, err := db.CreateRecipe(userA, recipe)
		if err != nil {
			t.Fatalf("failed to create recipe: %v", err)
		}

		got, err := db.GetRecipe(userA, id)
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
		_, err := db.GetRecipe(userA, "nonexistent-id")
		if err == nil {
			t.Error("expected error for non-existent recipe, got nil")
		}
	})

	t.Run("returns not-found when another user tries to read it", func(t *testing.T) {
		db := initRecipeDB(t)
		id, _ := db.CreateRecipe(userA, rec.Recipe{Name: "A's recipe", Instructions: rec.InstructionList{"Step"}})
		if _, err := db.GetRecipe(userB, id); err == nil {
			t.Error("expected not-found when user B reads user A's recipe, got nil")
		}
	})
}

func TestGetAllRecipes(t *testing.T) {
	t.Run("returns empty slice when no recipes exist", func(t *testing.T) {
		db := initRecipeDB(t)
		recipes, err := db.GetAllRecipes(userA)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(recipes) != 0 {
			t.Errorf("expected 0 recipes, got %d", len(recipes))
		}
	})

	t.Run("returns all created recipes", func(t *testing.T) {
		db := initRecipeDB(t)
		db.CreateRecipe(userA, rec.Recipe{Name: "Recipe 1", Instructions: rec.InstructionList{"Step 1"}})
		db.CreateRecipe(userA, rec.Recipe{Name: "Recipe 2", Instructions: rec.InstructionList{"Step 1"}})
		db.CreateRecipe(userA, rec.Recipe{Name: "Recipe 3", Instructions: rec.InstructionList{"Step 1"}})

		recipes, err := db.GetAllRecipes(userA)
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
		id, err := db.CreateRecipe(userA, rec.Recipe{
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
		if err := db.UpdateRecipe(userA, id, updated); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, err := db.GetRecipe(userA, id)
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
		err := db.UpdateRecipe(userA, "nonexistent-id", rec.Recipe{Name: "Ghost", Instructions: rec.InstructionList{"Step"}})
		if err == nil {
			t.Error("expected error for non-existent recipe, got nil")
		}
	})

	t.Run("returns not-found when another user tries to update it", func(t *testing.T) {
		db := initRecipeDB(t)
		id, _ := db.CreateRecipe(userA, rec.Recipe{
			Name:         "A's recipe",
			Instructions: rec.InstructionList{"Step"},
			Ingredients:  []rec.Ingredient{{Name: "salt", Quantity: "1"}},
		})
		err := db.UpdateRecipe(userB, id, rec.Recipe{
			Name:         "hijacked",
			Instructions: rec.InstructionList{"New step"},
			Ingredients:  []rec.Ingredient{{Name: "poison", Quantity: "1"}},
		})
		if err == nil {
			t.Fatal("expected not-found when user B updates user A's recipe, got nil")
		}
		got, _ := db.GetRecipe(userA, id)
		if got.Name != "A's recipe" {
			t.Errorf("recipe was modified by non-owner: got name %q", got.Name)
		}
	})
}

func TestDeleteRecipe(t *testing.T) {
	t.Run("deletes an existing recipe", func(t *testing.T) {
		db := initRecipeDB(t)
		id, err := db.CreateRecipe(userA, rec.Recipe{
			Name:         "To Delete",
			Instructions: rec.InstructionList{"Step 1"},
			Ingredients:  []rec.Ingredient{{Name: "sugar", Quantity: "1 cup"}},
		})
		if err != nil {
			t.Fatalf("failed to create recipe: %v", err)
		}

		if err := db.DeleteRecipe(userA, id); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = db.GetRecipe(userA, id)
		if err == nil {
			t.Error("expected error fetching deleted recipe, got nil")
		}
	})

	t.Run("returns error for non-existent recipe", func(t *testing.T) {
		db := initRecipeDB(t)
		err := db.DeleteRecipe(userA, "nonexistent-id")
		if err == nil {
			t.Error("expected error for non-existent recipe, got nil")
		}
	})

	t.Run("returns not-found when another user tries to delete it", func(t *testing.T) {
		db := initRecipeDB(t)
		id, _ := db.CreateRecipe(userA, rec.Recipe{Name: "A's recipe", Instructions: rec.InstructionList{"Step"}})
		if err := db.DeleteRecipe(userB, id); err == nil {
			t.Fatal("expected not-found when user B deletes user A's recipe, got nil")
		}
		if _, err := db.GetRecipe(userA, id); err != nil {
			t.Errorf("recipe was deleted by non-owner: %v", err)
		}
	})
}

func TestDeleteMealPlan(t *testing.T) {
	t.Run("deletes an existing meal plan", func(t *testing.T) {
		db := initRecipeDB(t)
		mpID, err := db.CreateMealPlan(userA, nil)
		if err != nil {
			t.Fatalf("failed to create meal plan: %v", err)
		}
		if err := db.DeleteMealPlan(userA, mpID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns not-found when another user tries to delete it", func(t *testing.T) {
		db := initRecipeDB(t)
		mpID, _ := db.CreateMealPlan(userA, nil)
		if err := db.DeleteMealPlan(userB, mpID); err == nil {
			t.Fatal("expected not-found when user B deletes user A's meal plan, got nil")
		}
		// Still deletable by owner — proves it wasn't actually removed.
		if err := db.DeleteMealPlan(userA, mpID); err != nil {
			t.Errorf("meal plan was deleted by non-owner: %v", err)
		}
	})
}

func TestAddRecipesToMealPlan(t *testing.T) {
	t.Run("returns not-found when another user tries to add to it", func(t *testing.T) {
		db := initRecipeDB(t)
		mpID, _ := db.CreateMealPlan(userA, nil)
		rID, _ := db.CreateRecipe(userB, rec.Recipe{Name: "B's recipe", Instructions: rec.InstructionList{"Step"}})
		if err := db.AddRecipesToMealPlan(userB, mpID, []string{rID}); err == nil {
			t.Fatal("expected not-found when user B adds recipes to user A's meal plan, got nil")
		}
	})
}

func TestGetGroceryListAuth(t *testing.T) {
	t.Run("returns not-found when another user tries to view it", func(t *testing.T) {
		db := initRecipeDB(t)
		mpID, _ := db.CreateMealPlan(userA, nil)
		if _, err := db.GetGroceryList(userB, mpID); err == nil {
			t.Error("expected not-found when user B reads grocery list for user A's meal plan, got nil")
		}
	})
}

func TestGroceryListByID(t *testing.T) {
	t.Run("retrieves owner's grocery list", func(t *testing.T) {
		db := initRecipeDB(t)
		id, err := db.CreateGroceryList(userA, "shop", []rec.GroceryListItem{{Name: "milk", Quantity: "1"}}, nil)
		if err != nil {
			t.Fatalf("failed to create grocery list: %v", err)
		}
		got, err := db.GetGroceryListByID(userA, id)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name != "shop" {
			t.Errorf("expected name %q, got %q", "shop", got.Name)
		}
	})

	t.Run("returns not-found when another user tries to read it", func(t *testing.T) {
		db := initRecipeDB(t)
		id, _ := db.CreateGroceryList(userA, "shop", nil, nil)
		if _, err := db.GetGroceryListByID(userB, id); err == nil {
			t.Error("expected not-found when user B reads user A's grocery list, got nil")
		}
	})
}

func TestUpdateGroceryListAuth(t *testing.T) {
	t.Run("returns not-found when another user tries to update it", func(t *testing.T) {
		db := initRecipeDB(t)
		id, _ := db.CreateGroceryList(userA, "shop", []rec.GroceryListItem{{Name: "milk", Quantity: "1"}}, nil)
		err := db.UpdateGroceryList(userB, id, []rec.GroceryListItem{{Name: "hijacked", Quantity: "9"}})
		if err == nil {
			t.Fatal("expected not-found when user B updates user A's grocery list, got nil")
		}
		got, _ := db.GetGroceryListByID(userA, id)
		if len(got.Items) != 1 || got.Items[0].Name != "milk" {
			t.Errorf("grocery list items were modified by non-owner: %+v", got.Items)
		}
	})
}

func TestDeleteGroceryListAuth(t *testing.T) {
	t.Run("returns not-found when another user tries to delete it", func(t *testing.T) {
		db := initRecipeDB(t)
		id, _ := db.CreateGroceryList(userA, "shop", nil, nil)
		if err := db.DeleteGroceryList(userB, id); err == nil {
			t.Fatal("expected not-found when user B deletes user A's grocery list, got nil")
		}
		if _, err := db.GetGroceryListByID(userA, id); err != nil {
			t.Errorf("grocery list was deleted by non-owner: %v", err)
		}
	})
}
