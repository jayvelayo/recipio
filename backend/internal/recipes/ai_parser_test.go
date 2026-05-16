package recipes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestParseRecipeText(t *testing.T) {
	want := Recipe{
		Name:        "French Toast",
		Description: "A classic breakfast dish",
		Ingredients: []Ingredient{
			{Name: "eggs", Quantity: "2"},
			{Name: "milk", Quantity: "1 cup"},
		},
		Instructions: InstructionList{
			"Crack open the eggs",
			"Add milk and egg into a bowl",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inner, _ := json.Marshal(want)
		enc := json.NewEncoder(w)
		enc.Encode(llmResponse{Response: string(inner), Done: false})
		enc.Encode(llmResponse{Response: "", Done: true})
	}))
	defer server.Close()

	parser := AIParser{APIURL: server.URL, Client: server.Client()}
	got, err := parser.ParseRecipeText("French toast recipe text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Name != want.Name {
		t.Errorf("Name: got %q, want %q", got.Name, want.Name)
	}
	if got.Description != want.Description {
		t.Errorf("Description: got %q, want %q", got.Description, want.Description)
	}
	if len(got.Ingredients) != len(want.Ingredients) {
		t.Errorf("Ingredients: got %d, want %d", len(got.Ingredients), len(want.Ingredients))
	}
	if len(got.Instructions) != len(want.Instructions) {
		t.Errorf("Instructions: got %d, want %d", len(got.Instructions), len(want.Instructions))
	}
}

func TestParseRecipeTextLive(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run against live Ollama API")
	}

	parser := NewAIParser()
	recipe, err := parser.ParseRecipeText(`French toast
2 eggs
1 cup milk
3 tsp parsley, minced

-. Crack open the egg
-. Add milk and egg into a bowl
-. Add parsley if desired`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if recipe.Name == "" {
		t.Error("expected non-empty recipe name")
	}
	if len(recipe.Ingredients) == 0 {
		t.Error("expected at least one ingredient")
	}
	if len(recipe.Instructions) == 0 {
		t.Error("expected at least one instruction")
	}

	t.Logf("Parsed recipe: %+v", recipe)
}
