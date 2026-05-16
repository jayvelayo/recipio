package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	rec "github.com/jayvelayo/recipio/internal/recipes"
)

func handleDesignCreateRecipe(recipeDb rec.RecipeDatabase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}
		var recipe rec.Recipe
		if err := json.NewDecoder(r.Body).Decode(&recipe); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}
		recipe.Name = sanitizeRecipeText(recipe.Name)
		var ings []rec.Ingredient
		for _, ing := range recipe.Ingredients {
			if name := sanitizeRecipeText(ing.Name); name != "" {
				ings = append(ings, rec.Ingredient{Name: name, Quantity: sanitizeRecipeText(ing.Quantity)})
			}
		}
		recipe.Ingredients = ings
		var instructions rec.InstructionList
		for _, s := range recipe.Instructions {
			if s := sanitizeRecipeText(s); s != "" {
				instructions = append(instructions, s)
			}
		}
		recipe.Instructions = instructions
		if recipe.Name == "" || len(recipe.Ingredients) == 0 || len(recipe.Instructions) == 0 {
			http.Error(w, "Missing required fields: name, ingredients, instructions", http.StatusBadRequest)
			return
		}
		recipeID, err := recipeDb.CreateRecipe(recipe)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		encodeJson(w, http.StatusCreated, CreateRecipeResponse{
			ID:      strconv.FormatUint(recipeID, 10),
			Message: "Recipe created successfully",
		})
	})
}

func handleDesignDeleteRecipe(recipeDb rec.RecipeDatabase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil || id < 1 {
			http.Error(w, "Invalid recipe id", http.StatusBadRequest)
			return
		}
		err = recipeDb.DeleteRecipe(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

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
		encodeJson(w, http.StatusOK, recipe)
	})
}

func handleDesignGetAllRecipes(recipeDb rec.RecipeDatabase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recipes, err := recipeDb.GetAllRecipes()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		encodeJson(w, http.StatusOK, recipes)
	})
}
