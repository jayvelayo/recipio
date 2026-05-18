package recipes

type InstructionList []string

type Ingredient struct {
	Name     string `json:"name"`
	Quantity string `json:"quantity"`
}

type Recipe struct {
	ID           string          `json:"id" deep:"-"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Ingredients  []Ingredient    `json:"ingredients"`
	Instructions InstructionList `json:"instructions"`
	Tags         []string        `json:"tags"`
}

type Recipes []Recipe

type MealPlanSummary struct {
	ID          string   `json:"id"`
	RecipeNames []string `json:"recipe_names"`
}

type IngredientSource struct {
	Quantity   string `json:"quantity"`
	RecipeName string `json:"recipe_name"`
}

type GroceryListItem struct {
	Name     string             `json:"name"`
	Quantity string             `json:"quantity"`
	Checked  bool               `json:"checked"`
	Sources  []IngredientSource `json:"sources,omitempty"`
}

type GroceryList struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Items      []GroceryListItem `json:"items"`
	MealPlanID *string           `json:"meal_plan_id,omitempty"`
}
