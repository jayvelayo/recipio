package recipes

type InstructionList []string

type Ingredient struct {
	Name     string `json:"name"`
	Quantity string `json:"quantity"`
}

type Recipe struct {
	ID           int             `json:"id" deep:"-"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Ingredients  []Ingredient    `json:"ingredients"`
	Instructions InstructionList `json:"instructions"`
}

type Recipes []Recipe

// MealPlanSummary describes a meal plan for list responses.
type MealPlanSummary struct {
	ID          string   `json:"id"`
	RecipeNames []string `json:"recipe_names"`
}

// GroceryListItem represents an item in a grocery list
type GroceryListItem struct {
	Name     string `json:"name"`
	Quantity string `json:"quantity"`
	Checked  bool   `json:"checked"`
}

// GroceryList represents a grocery list
type GroceryList struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Items      []GroceryListItem `json:"items"`
	MealPlanID *string           `json:"meal_plan_id,omitempty"` // optional, if created from meal plan
}

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
