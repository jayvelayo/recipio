package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	rec "github.com/jayvelayo/recipio/internal/recipes"
)

// designRecipeToInternal converts API request to internal recipe model
func designRecipeToInternal(body designRecipeRequest) rec.Recipe {
	var recipe rec.Recipe
	recipe.Name = body.Name
	recipe.Instructions = rec.InstructionList(body.Steps)
	for _, line := range body.Ingredients {
		words := strings.Fields(line)
		if len(words) == 0 {
			continue
		}
		recipe.Ingredients = append(recipe.Ingredients, rec.Ingredient{
			Name:     words[len(words)-1],
			Quantity: strings.Join(words[0:len(words)-1], " "),
		})
	}
	return recipe
}

// internalRecipeToDesign converts internal recipe model to API response
func internalRecipeToDesign(recipe rec.Recipe) designRecipeResponse {
	idStr := strconv.Itoa(recipe.ID)
	ingStrings := make([]string, 0, len(recipe.Ingredients))
	for _, ing := range recipe.Ingredients {
		s := strings.TrimSpace(ing.Quantity) + " " + strings.TrimSpace(ing.Name)
		ingStrings = append(ingStrings, strings.TrimSpace(s))
	}
	return designRecipeResponse{
		ID:          idStr,
		Name:        recipe.Name,
		Ingredients: ingStrings,
		Steps:       recipe.Instructions,
	}
}

// handleDesignCreateRecipe creates a new recipe (Design API)
func handleDesignCreateRecipe(recipeDb rec.RecipeDatabase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}
		var body designRecipeRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}
		recipe := designRecipeToInternal(body)
		if recipe.Name == "" || len(recipe.Ingredients) == 0 || len(recipe.Instructions) == 0 {
			http.Error(w, "Missing required fields: name, ingredients, steps", http.StatusBadRequest)
			return
		}
		recipeID, err := recipeDb.CreateRecipe(recipe)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		encodeJson(w, http.StatusCreated, designCreateRecipeResponse{
			ID:      strconv.FormatUint(recipeID, 10),
			Message: "Recipe created successfully",
		})
	})
}

// handleDesignGetRecipe retrieves a single recipe (Design API)
func handleDesignGetRecipe(recipeDb rec.RecipeDatabase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil || id < 1 {
			http.Error(w, "Invalid recipe id", http.StatusBadRequest)
			return
		}
		recipe, err := recipeDb.GetRecipe(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		encodeJson(w, http.StatusOK, internalRecipeToDesign(recipe))
	})
}

// handleDesignGetAllRecipes retrieves all recipes (Design API)
func handleDesignGetAllRecipes(recipeDb rec.RecipeDatabase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recipes, err := recipeDb.GetAllRecipes()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list := make([]designRecipeResponse, 0, len(recipes))
		for _, recipe := range recipes {
			list = append(list, internalRecipeToDesign(recipe))
		}
		encodeJson(w, http.StatusOK, list)
	})
}
