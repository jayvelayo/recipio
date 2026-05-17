package recipes

import (
	"fmt"
	"strconv"
)

type MockRecipeDatabase struct {
	recipes       []Recipe
	recipeCount   int
	mealPlans     []MealPlanSummary
	mealPlanCount int
	groceryLists  []GroceryList
	groceryCount  int
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
	db.mealPlanCount++
	id := strconv.Itoa(db.mealPlanCount)
	db.mealPlans = append(db.mealPlans, MealPlanSummary{ID: id, RecipeNames: recipeIDs})
	return id, nil
}

func (db *MockRecipeDatabase) GetAllMealPlans() ([]MealPlanSummary, error) {
	return db.mealPlans, nil
}

func (db *MockRecipeDatabase) GetGroceryList(mealPlanID string) ([]string, error) {
	for _, plan := range db.mealPlans {
		if plan.ID == mealPlanID {
			return []string{"mock ingredient"}, nil
		}
	}
	return nil, fmt.Errorf("meal plan not found")
}

func (db *MockRecipeDatabase) DeleteMealPlan(mealPlanID string) error {
	for i, plan := range db.mealPlans {
		if plan.ID == mealPlanID {
			db.mealPlans = append(db.mealPlans[:i], db.mealPlans[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("meal plan not found")
}

func (db *MockRecipeDatabase) CreateGroceryList(name string, items []GroceryListItem, mealPlanID *string) (string, error) {
	db.groceryCount++
	id := strconv.Itoa(db.groceryCount)
	db.groceryLists = append(db.groceryLists, GroceryList{ID: id, Name: name, Items: items, MealPlanID: mealPlanID})
	return id, nil
}

func (db *MockRecipeDatabase) GetAllGroceryLists() ([]GroceryList, error) {
	return db.groceryLists, nil
}

func (db *MockRecipeDatabase) GetGroceryListByID(id string) (GroceryList, error) {
	for _, list := range db.groceryLists {
		if list.ID == id {
			return list, nil
		}
	}
	return GroceryList{}, fmt.Errorf("grocery list not found")
}

func (db *MockRecipeDatabase) UpdateGroceryList(id string, items []GroceryListItem) error {
	for i, list := range db.groceryLists {
		if list.ID == id {
			db.groceryLists[i].Items = items
			return nil
		}
	}
	return fmt.Errorf("grocery list not found")
}

func (db *MockRecipeDatabase) DeleteGroceryList(id string) error {
	for i, list := range db.groceryLists {
		if list.ID == id {
			db.groceryLists = append(db.groceryLists[:i], db.groceryLists[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("grocery list not found")
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
