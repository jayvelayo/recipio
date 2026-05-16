package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	rec "github.com/jayvelayo/recipio/internal/recipes"
)

func handleDesignParseRecipe(parser rec.AIParser) http.Handler {
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

		recipe, err := parser.ParseRecipeText(sanitizeRecipeText(body.RawRecipeText))
		if err != nil {
			if errors.Is(err, rec.ErrLLMTimeout) {
				http.Error(w, "LLM request timed out", http.StatusGatewayTimeout)
				return
			}
			log.Printf("parse recipe error: %v", err)
			http.Error(w, "Failed to parse recipe", http.StatusInternalServerError)
			return
		}

		encodeJson(w, http.StatusOK, internalRecipeToDesign(recipe))
	})
}
