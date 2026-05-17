package recipes

type RecipeDatabase interface {
	CreateRecipe(userID string, recipe Recipe) (string, error)
	GetRecipe(id string) (Recipe, error)
	GetAllRecipes(userID string) (Recipes, error)
	UpdateRecipe(id string, recipe Recipe) error
	DeleteRecipe(id string) error
	AddRecipeToMealPlan(id string) error
	CreateMealPlan(userID string, recipeIDs []string) (mealPlanID string, err error)
	GetAllMealPlans(userID string) ([]MealPlanSummary, error)
	GetGroceryList(mealPlanID string) ([]Ingredient, error)
	DeleteMealPlan(mealPlanID string) error
	CreateGroceryList(userID string, name string, items []GroceryListItem, mealPlanID *string) (string, error)
	GetAllGroceryLists(userID string) ([]GroceryList, error)
	GetGroceryListByID(id string) (GroceryList, error)
	UpdateGroceryList(id string, items []GroceryListItem) error
	DeleteGroceryList(id string) error
	CloseDb()
}
