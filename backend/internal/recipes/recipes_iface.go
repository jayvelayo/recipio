package recipes

type RecipeDatabase interface {
	CreateRecipe(userID string, recipe Recipe) (string, error)
	GetRecipe(userID, id string) (Recipe, error)
	GetAllRecipes(userID string) (Recipes, error)
	UpdateRecipe(userID, id string, recipe Recipe) error
	DeleteRecipe(userID, id string) error
	AddRecipeToMealPlan(id string) error
	CreateMealPlan(userID string, recipeIDs []string) (mealPlanID string, err error)
	AddRecipesToMealPlan(userID, mealPlanID string, recipeIDs []string) error
	GetAllMealPlans(userID string) ([]MealPlanSummary, error)
	GetGroceryList(userID, mealPlanID string) ([]GroceryListItem, error)
	DeleteMealPlan(userID, mealPlanID string) error
	CreateGroceryList(userID string, name string, items []GroceryListItem, mealPlanID *string) (string, error)
	GetAllGroceryLists(userID string) ([]GroceryList, error)
	GetGroceryListByID(userID, id string) (GroceryList, error)
	UpdateGroceryList(userID, id string, items []GroceryListItem) error
	DeleteGroceryList(userID, id string) error
	CloseDb()
}
