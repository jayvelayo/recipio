package recipes

type RecipeDatabase interface {
	CreateRecipe(recipe Recipe) (uint64, error)
	GetRecipe(id int) (Recipe, error)
	GetAllRecipes() (Recipes, error)
	DeleteRecipe(id int) error
	AddRecipeToMealPlan(id int) error
	CreateMealPlan(recipeIDs []string) (mealPlanID string, err error)
	GetAllMealPlans() ([]MealPlanSummary, error)
	GetGroceryList(mealPlanID string) (ingredients []string, err error)
	DeleteMealPlan(mealPlanID string) error
	CreateGroceryList(name string, items []GroceryListItem, mealPlanID *string) (string, error)
	GetAllGroceryLists() ([]GroceryList, error)
	GetGroceryListByID(id string) (GroceryList, error)
	UpdateGroceryList(id string, items []GroceryListItem) error
	DeleteGroceryList(id string) error
	CloseDb()
}
