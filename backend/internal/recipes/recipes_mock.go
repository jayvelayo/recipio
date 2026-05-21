package recipes

import (
	"fmt"
	"strconv"
)

type MockRecipeDatabase struct {
	recipes        []Recipe
	recipeOwner    map[string]string
	nextID         int
	mealPlans      []MealPlanSummary
	mealPlanOwner  map[string]string
	mealPlanCount  int
	groceryLists   []GroceryList
	groceryOwner   map[string]string
	groceryCount   int
}

func (db *MockRecipeDatabase) ensureMaps() {
	if db.recipeOwner == nil {
		db.recipeOwner = map[string]string{}
	}
	if db.mealPlanOwner == nil {
		db.mealPlanOwner = map[string]string{}
	}
	if db.groceryOwner == nil {
		db.groceryOwner = map[string]string{}
	}
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
	db.ensureMaps()
	db.nextID++
	id := strconv.Itoa(db.nextID)
	newRecipe.ID = id
	db.recipes = append(db.recipes, newRecipe)
	db.recipeOwner[id] = userID
	return id, nil
}

func (db *MockRecipeDatabase) GetRecipeCount() int {
	return len(db.recipes)
}

func (db *MockRecipeDatabase) GetRecipe(userID, id string) (Recipe, error) {
	if db.recipeOwner[id] != userID {
		return Recipe{}, fmt.Errorf("recipe not found")
	}
	for _, r := range db.recipes {
		if r.ID == id {
			return r, nil
		}
	}
	return Recipe{}, fmt.Errorf("recipe not found")
}

func (db *MockRecipeDatabase) GetAllRecipes(userID string) (Recipes, error) {
	var out Recipes
	for _, r := range db.recipes {
		if db.recipeOwner[r.ID] == userID {
			out = append(out, r)
		}
	}
	return out, nil
}

func (db *MockRecipeDatabase) AddRecipeToMealPlan(id string) error {
	return nil
}

func (db *MockRecipeDatabase) CreateMealPlan(userID string, recipeIDs []string) (string, error) {
	db.ensureMaps()
	db.mealPlanCount++
	id := strconv.Itoa(db.mealPlanCount)
	db.mealPlans = append(db.mealPlans, MealPlanSummary{ID: id, RecipeNames: recipeIDs})
	db.mealPlanOwner[id] = userID
	return id, nil
}

func (db *MockRecipeDatabase) AddRecipesToMealPlan(userID, mealPlanID string, recipeIDs []string) error {
	if db.mealPlanOwner[mealPlanID] != userID {
		return fmt.Errorf("meal plan not found")
	}
	for _, plan := range db.mealPlans {
		if plan.ID == mealPlanID {
			plan.RecipeNames = append(plan.RecipeNames, recipeIDs...)
			return nil
		}
	}
	return fmt.Errorf("meal plan not found")
}

func (db *MockRecipeDatabase) GetAllMealPlans(userID string) ([]MealPlanSummary, error) {
	var out []MealPlanSummary
	for _, p := range db.mealPlans {
		if db.mealPlanOwner[p.ID] == userID {
			out = append(out, p)
		}
	}
	return out, nil
}

func (db *MockRecipeDatabase) GetGroceryList(userID, mealPlanID string) ([]GroceryListItem, error) {
	if db.mealPlanOwner[mealPlanID] != userID {
		return nil, fmt.Errorf("meal plan not found")
	}
	for _, plan := range db.mealPlans {
		if plan.ID == mealPlanID {
			return []GroceryListItem{{
				Name:     "mock ingredient",
				Category: UncategorizedName,
				Sources:  []IngredientSource{{Quantity: "1", RecipeName: "mock recipe"}},
			}}, nil
		}
	}
	return nil, fmt.Errorf("meal plan not found")
}

func (db *MockRecipeDatabase) DeleteMealPlan(userID, mealPlanID string) error {
	if db.mealPlanOwner[mealPlanID] != userID {
		return fmt.Errorf("meal plan not found")
	}
	for i, plan := range db.mealPlans {
		if plan.ID == mealPlanID {
			db.mealPlans = append(db.mealPlans[:i], db.mealPlans[i+1:]...)
			delete(db.mealPlanOwner, mealPlanID)
			return nil
		}
	}
	return fmt.Errorf("meal plan not found")
}

func (db *MockRecipeDatabase) CreateGroceryList(userID string, name string, items []GroceryListItem, mealPlanID *string) (string, error) {
	db.ensureMaps()
	db.groceryCount++
	id := strconv.Itoa(db.groceryCount)
	db.groceryLists = append(db.groceryLists, GroceryList{ID: id, Name: name, Items: items, MealPlanID: mealPlanID})
	db.groceryOwner[id] = userID
	return id, nil
}

func (db *MockRecipeDatabase) GetAllGroceryLists(userID string) ([]GroceryList, error) {
	var out []GroceryList
	for _, list := range db.groceryLists {
		if db.groceryOwner[list.ID] == userID {
			out = append(out, list)
		}
	}
	return out, nil
}

func (db *MockRecipeDatabase) GetGroceryListByID(userID, id string) (GroceryList, error) {
	if db.groceryOwner[id] != userID {
		return GroceryList{}, fmt.Errorf("grocery list not found")
	}
	for _, list := range db.groceryLists {
		if list.ID == id {
			return list, nil
		}
	}
	return GroceryList{}, fmt.Errorf("grocery list not found")
}

func (db *MockRecipeDatabase) UpdateGroceryList(userID, id string, items []GroceryListItem) error {
	if db.groceryOwner[id] != userID {
		return fmt.Errorf("grocery list not found")
	}
	for i, list := range db.groceryLists {
		if list.ID == id {
			db.groceryLists[i].Items = items
			return nil
		}
	}
	return fmt.Errorf("grocery list not found")
}

func (db *MockRecipeDatabase) DeleteGroceryList(userID, id string) error {
	if db.groceryOwner[id] != userID {
		return fmt.Errorf("grocery list not found")
	}
	for i, list := range db.groceryLists {
		if list.ID == id {
			db.groceryLists = append(db.groceryLists[:i], db.groceryLists[i+1:]...)
			delete(db.groceryOwner, id)
			return nil
		}
	}
	return fmt.Errorf("grocery list not found")
}

func (db *MockRecipeDatabase) UpdateRecipe(userID, id string, recipe Recipe) error {
	if db.recipeOwner[id] != userID {
		return fmt.Errorf("recipe not found")
	}
	for i, r := range db.recipes {
		if r.ID == id {
			recipe.ID = id
			db.recipes[i] = recipe
			return nil
		}
	}
	return fmt.Errorf("recipe not found")
}

func (db *MockRecipeDatabase) DeleteRecipe(userID, id string) error {
	if db.recipeOwner[id] != userID {
		return fmt.Errorf("recipe not found")
	}
	for i, r := range db.recipes {
		if r.ID == id {
			db.recipes = append(db.recipes[:i], db.recipes[i+1:]...)
			delete(db.recipeOwner, id)
			return nil
		}
	}
	return fmt.Errorf("recipe not found")
}

func (db *MockRecipeDatabase) CloseDb() {
}
