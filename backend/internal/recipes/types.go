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

type GroceryListItem struct {
	Name     string `json:"name"`
	Quantity string `json:"quantity"`
	Checked  bool   `json:"checked"`
}

type GroceryList struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Items      []GroceryListItem `json:"items"`
	MealPlanID *string           `json:"meal_plan_id,omitempty"`
}
