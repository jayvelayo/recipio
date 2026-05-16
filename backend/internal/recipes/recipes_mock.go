package recipes

import "fmt"

type MockRecipeDatabase struct {
	recipes     []Recipe
	recipeCount int
}

func (db *MockRecipeDatabase) CreateRecipe(newRecipe Recipe) (uint64, error) {
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

func (db *MockRecipeDatabase) GetRecipeCount() int {
	return db.recipeCount
}

func (db *MockRecipeDatabase) GetRecipe(id int) (Recipe, error) {
	if id < 1 || id > db.recipeCount {
		return Recipe{}, fmt.Errorf("recipe not found")
	}
	return db.recipes[id-1], nil
}

func (db *MockRecipeDatabase) GetAllRecipes() (Recipes, error) {
	return db.recipes, nil
}

func (db *MockRecipeDatabase) AddRecipeToMealPlan(id int) error {
	return nil
}

func (db *MockRecipeDatabase) CreateMealPlan(recipeIDs []string) (string, error) {
	return "1", nil
}

func (db *MockRecipeDatabase) GetAllMealPlans() ([]MealPlanSummary, error) {
	return nil, nil
}

func (db *MockRecipeDatabase) GetGroceryList(mealPlanID string) ([]string, error) {
	return []string{"mock ingredient"}, nil
}

func (db *MockRecipeDatabase) DeleteMealPlan(mealPlanID string) error {
	return nil
}

func (db *MockRecipeDatabase) CreateGroceryList(name string, items []GroceryListItem, mealPlanID *string) (string, error) {
	return "1", nil
}

func (db *MockRecipeDatabase) GetAllGroceryLists() ([]GroceryList, error) {
	return []GroceryList{}, nil
}

func (db *MockRecipeDatabase) GetGroceryListByID(id string) (GroceryList, error) {
	return GroceryList{}, fmt.Errorf("not implemented")
}

func (db *MockRecipeDatabase) UpdateGroceryList(id string, items []GroceryListItem) error {
	return nil
}

func (db *MockRecipeDatabase) DeleteGroceryList(id string) error {
	return nil
}

func (db *MockRecipeDatabase) UpdateRecipe(id int, recipe Recipe) error {
	if id < 1 || id > db.recipeCount {
		return fmt.Errorf("recipe not found")
	}
	recipe.ID = id
	db.recipes[id-1] = recipe
	return nil
}

func (db *MockRecipeDatabase) DeleteRecipe(id int) error {
	if id < 1 || id > db.recipeCount {
		return fmt.Errorf("recipe not found")
	}
	// Remove the recipe from the slice
	db.recipes = append(db.recipes[:id-1], db.recipes[id:]...)
	db.recipeCount--
	return nil
}

func (db *MockRecipeDatabase) CloseDb() {
}
