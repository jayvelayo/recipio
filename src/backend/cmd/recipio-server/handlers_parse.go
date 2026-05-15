package main

import (
	"encoding/json"
	"net/http"
)

// handleDesignParseRecipe parses raw recipe text using AI (Design API)
func handleDesignParseRecipe() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}
		var body struct {
			RawRecipeText string `json:"raw_recipe_text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}
		if body.RawRecipeText == "" {
			http.Error(w, "Raw recipe text is required", http.StatusBadRequest)
			return
		}

		// TODO: Implement actual AI parsing logic here
		// For now, return a dummy response for demonstration
		response := designRecipeResponse{
			ID:          "parsed-recipe-1",
			Name:        "Parsed Recipe",
			Ingredients: []string{"1 cup flour", "2 eggs", "1/2 cup sugar"},
			Steps:       []string{"Mix ingredients", "Bake for 30 minutes"},
		}
		encodeJson(w, http.StatusOK, response)
	})
}
