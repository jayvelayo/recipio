package recipes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Prompter is the LLM capability that recipes needs: send a prompt string,
// get a completion string back. Implementations live outside this package
// (e.g. internal/llm.GroqClient) so recipes stays free of HTTP / vendor types.
type Prompter interface {
	FetchPrompt(ctx context.Context, prompt string) (string, error)
}

func ParseRecipeText(p Prompter, text string) (Recipe, error) {
	prompt := fmt.Sprintf(`Extract the recipe into JSON with this schema:
{
  "name": string,
  "description": string,
  "ingredients": [{"name": string, "quantity": string}],
  "instructions": [string],
  "tags": [string]
}

For "tags", infer 1-3 short lowercase tags describing the cuisine or diet
(e.g. "italian", "vegetarian", "thai", "gluten-free"). Omit if unclear.

Recipe:
%s

Return ONLY valid JSON.`, text)

	content, err := p.FetchPrompt(context.Background(), prompt)
	if err != nil {
		return Recipe{}, err
	}
	var recipe Recipe
	if err := json.Unmarshal([]byte(content), &recipe); err != nil {
		return Recipe{}, fmt.Errorf("parse recipe json: %w", err)
	}
	return recipe, nil
}

// CategorizeIngredients asks the model to place each ingredient name into one of the
// allowed categories. Returns a map keyed by the input names (caller decides the
// normalization scheme). Names the model can't place are simply absent from the result;
// the caller is responsible for falling back to Uncategorized.
func CategorizeIngredients(p Prompter, ctx context.Context, names []string, categories []string) (map[string]string, error) {
	if len(names) == 0 {
		return map[string]string{}, nil
	}
	allowed := strings.Join(categories, ", ")
	listed := strings.Join(names, "\n")
	prompt := fmt.Sprintf(`Categorize each grocery ingredient into exactly one of these categories:
%s

Return JSON with the schema:
{ "categories": { "<ingredient-name>": "<category>", ... } }

Use the ingredient names exactly as given. Pick the single best category from the
list above for each ingredient. If none fit, use "%s".

Ingredients:
%s

Return ONLY valid JSON.`, allowed, UncategorizedName, listed)

	content, err := p.FetchPrompt(ctx, prompt)
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Categories map[string]string `json:"categories"`
	}
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return nil, fmt.Errorf("parse categorize response: %w", err)
	}
	valid := make(map[string]struct{}, len(categories))
	for _, c := range categories {
		valid[c] = struct{}{}
	}
	result := make(map[string]string, len(parsed.Categories))
	for name, cat := range parsed.Categories {
		if _, ok := valid[cat]; ok {
			result[name] = cat
		}
	}
	return result, nil
}
