package recipes

import (
	"fmt"
	"strconv"
)

type MockRecipeDatabase struct {
	recipes      []Recipe
	nextID       int
	mealPlans    []MealPlanSummary
	mealPlanCount int
	groceryLists []GroceryList
	groceryCount int
}

func (db *MockRecipeDatabase) CreateRecipe(userID string, newRecipe Recipe) (string, error) {
	if newRecipe.Name == "" {
		return "", fmt.Errorf("recipe name cannot be empty")
	}
	for _, recipe := range db.recipes {
		if newRecipe.Name == recipe.Name {
			return "", fmt.Errorf("recipe with this name already exists")
		}
	}
	db.nextID++
	id := strconv.Itoa(db.nextID)
	newRecipe.ID = id
	db.recipes = append(db.recipes, newRecipe)
	return id, nil
}

func (db *MockRecipeDatabase) GetRecipeCount() int {
	return len(db.recipes)
}

func (db *MockRecipeDatabase) GetRecipe(id string) (Recipe, error) {
	for _, r := range db.recipes {
		if r.ID == id {
			return r, nil
		}
	}
	return Recipe{}, fmt.Errorf("recipe not found")
}

func (db *MockRecipeDatabase) GetAllRecipes(userID string) (Recipes, error) {
	return db.recipes, nil
}

func (db *MockRecipeDatabase) AddRecipeToMealPlan(id string) error {
	return nil
}

func (db *MockRecipeDatabase) CreateMealPlan(userID string, recipeIDs []string) (string, error) {
	db.mealPlanCount++
	id := strconv.Itoa(db.mealPlanCount)
	db.mealPlans = append(db.mealPlans, MealPlanSummary{ID: id, RecipeNames: recipeIDs})
	return id, nil
}

func (db *MockRecipeDatabase) GetAllMealPlans(userID string) ([]MealPlanSummary, error) {
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

func (db *MockRecipeDatabase) CreateGroceryList(userID string, name string, items []GroceryListItem, mealPlanID *string) (string, error) {
	db.groceryCount++
	id := strconv.Itoa(db.groceryCount)
	db.groceryLists = append(db.groceryLists, GroceryList{ID: id, Name: name, Items: items, MealPlanID: mealPlanID})
	return id, nil
}

func (db *MockRecipeDatabase) GetAllGroceryLists(userID string) ([]GroceryList, error) {
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

func (db *MockRecipeDatabase) UpdateRecipe(id string, recipe Recipe) error {
	for i, r := range db.recipes {
		if r.ID == id {
			recipe.ID = id
			db.recipes[i] = recipe
			return nil
		}
	}
	return fmt.Errorf("recipe not found")
}

func (db *MockRecipeDatabase) DeleteRecipe(id string) error {
	for i, r := range db.recipes {
		if r.ID == id {
			db.recipes = append(db.recipes[:i], db.recipes[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("recipe not found")
}

func (db *MockRecipeDatabase) CloseDb() {
}
