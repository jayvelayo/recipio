package recipes

type RecipeDatabase interface {
	CreateRecipe(userID string, recipe Recipe) (uint64, error)
	GetRecipe(id int) (Recipe, error)
	GetAllRecipes(userID string) (Recipes, error)
	UpdateRecipe(id int, recipe Recipe) error
	DeleteRecipe(id int) error
	AddRecipeToMealPlan(id int) error
	CreateMealPlan(userID string, recipeIDs []string) (mealPlanID string, err error)
	GetAllMealPlans(userID string) ([]MealPlanSummary, error)
	GetGroceryList(mealPlanID string) (ingredients []string, err error)
	DeleteMealPlan(mealPlanID string) error
	CreateGroceryList(userID string, name string, items []GroceryListItem, mealPlanID *string) (string, error)
	GetAllGroceryLists(userID string) ([]GroceryList, error)
	GetGroceryListByID(id string) (GroceryList, error)
	UpdateGroceryList(id string, items []GroceryListItem) error
	DeleteGroceryList(id string) error
	CloseDb()
}
