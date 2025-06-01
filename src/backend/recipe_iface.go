package main

type instructionList []string

type Ingredient struct {
	Name     string `json:"name"`
	Quantity string `json:"quantity"`
}

type Recipe struct {
	ID           int             `json:"id" deep:"-"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Ingredients  []Ingredient    `json:"ingredients"`
	Instructions instructionList `json:"instructions"`
}

type Recipes []Recipe

type RecipeDatabase interface {
	createRecipe(recipe Recipe) (uint64, error)
	getRecipe(id int) (Recipe, error)
	getAllRecipes() (Recipes, error)
	closeDb()
}
