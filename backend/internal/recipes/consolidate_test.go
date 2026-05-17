package recipes

import (
	"testing"
)

func TestConsolidateIngredients(t *testing.T) {
	tests := []struct {
		name  string
		input []Ingredient
		want  []Ingredient
	}{
		{
			name:  "no ingredients",
			input: []Ingredient{},
			want:  []Ingredient{},
		},
		{
			name: "single ingredient unchanged",
			input: []Ingredient{
				{Name: "flour", Quantity: "2 cups"},
			},
			want: []Ingredient{
				{Name: "flour", Quantity: "2 cups"},
			},
		},
		{
			name: "deduplicate same unit",
			input: []Ingredient{
				{Name: "flour", Quantity: "1 cup"},
				{Name: "flour", Quantity: "1 cup"},
			},
			want: []Ingredient{
				{Name: "flour", Quantity: "2 cups"},
			},
		},
		{
			name: "convert cups and tbsp to cups",
			input: []Ingredient{
				{Name: "flour", Quantity: "1 cup"},
				{Name: "flour", Quantity: "2 tbsp"},
			},
			want: []Ingredient{
				{Name: "flour", Quantity: "1 1/8 cups"},
			},
		},
		{
			name: "convert tsp to tbsp",
			input: []Ingredient{
				{Name: "salt", Quantity: "1 tsp"},
				{Name: "salt", Quantity: "2 tsp"},
			},
			want: []Ingredient{
				{Name: "salt", Quantity: "1 tbsp"},
			},
		},
		{
			name: "imperial weight merging",
			input: []Ingredient{
				{Name: "chicken", Quantity: "8 oz"},
				{Name: "chicken", Quantity: "8 oz"},
			},
			want: []Ingredient{
				{Name: "chicken", Quantity: "1 lb"},
			},
		},
		{
			name: "metric weight merging",
			input: []Ingredient{
				{Name: "butter", Quantity: "100g"},
				{Name: "butter", Quantity: "150g"},
			},
			want: []Ingredient{
				{Name: "butter", Quantity: "250 g"},
			},
		},
		{
			name: "count merging with same unit",
			input: []Ingredient{
				{Name: "eggs", Quantity: "2"},
				{Name: "eggs", Quantity: "3"},
			},
			want: []Ingredient{
				{Name: "eggs", Quantity: "5"},
			},
		},
		{
			name: "incompatible units concatenate",
			input: []Ingredient{
				{Name: "flour", Quantity: "1 cup"},
				{Name: "flour", Quantity: "100g"},
			},
			want: []Ingredient{
				{Name: "flour", Quantity: "1 cup + 100g"},
			},
		},
		{
			name: "case-insensitive name matching",
			input: []Ingredient{
				{Name: "Butter", Quantity: "1/2 cup"},
				{Name: "butter", Quantity: "1/4 cup"},
			},
			want: []Ingredient{
				{Name: "Butter", Quantity: "3/4 cup"},
			},
		},
		{
			name: "distinct ingredients preserved separately",
			input: []Ingredient{
				{Name: "flour", Quantity: "2 cups"},
				{Name: "sugar", Quantity: "1 cup"},
			},
			want: []Ingredient{
				{Name: "flour", Quantity: "2 cups"},
				{Name: "sugar", Quantity: "1 cup"},
			},
		},
		{
			name: "mixed number parsing",
			input: []Ingredient{
				{Name: "milk", Quantity: "1 1/2 cups"},
				{Name: "milk", Quantity: "1/2 cup"},
			},
			want: []Ingredient{
				{Name: "milk", Quantity: "2 cups"},
			},
		},
		{
			name: "empty quantity treated as unknown",
			input: []Ingredient{
				{Name: "salt", Quantity: ""},
				{Name: "salt", Quantity: "1 tsp"},
			},
			want: []Ingredient{
				{Name: "salt", Quantity: "1 tsp"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConsolidateIngredients(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d ingredients, want %d: %+v", len(got), len(tt.want), got)
			}
			for i, g := range got {
				w := tt.want[i]
				if g.Name != w.Name || g.Quantity != w.Quantity {
					t.Errorf("ingredient[%d]: got {%q, %q}, want {%q, %q}", i, g.Name, g.Quantity, w.Name, w.Quantity)
				}
			}
		})
	}
}

func TestParseQty(t *testing.T) {
	tests := []struct {
		input  string
		system string
		amount float64
	}{
		{"2 cups", "volume", 96},
		{"1/2 cup", "volume", 24},
		{"1 1/2 tbsp", "volume", 4.5},
		{"3 tsp", "volume", 3},
		{"8 oz", "imperial_weight", 8},
		{"1 lb", "imperial_weight", 16},
		{"250g", "metric_weight", 250},
		{"1 kg", "metric_weight", 1000},
		{"500 ml", "metric_volume", 500},
		{"1 liter", "metric_volume", 1000},
		{"3 cans", "count:cans", 3},
		{"", "unknown", 0},
		{"to taste", "unknown", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseQty(tt.input)
			if got.system != tt.system {
				t.Errorf("system: got %q, want %q", got.system, tt.system)
			}
			if tt.system != "unknown" && got.amount != tt.amount {
				t.Errorf("amount: got %f, want %f", got.amount, tt.amount)
			}
		})
	}
}

func TestFormatAmount(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{1, "1"},
		{2, "2"},
		{0.5, "1/2"},
		{0.25, "1/4"},
		{0.75, "3/4"},
		{1.5, "1 1/2"},
		{2.25, "2 1/4"},
		{0.333, "1/3"},
	}
	for _, tt := range tests {
		got := formatAmount(tt.input)
		if got != tt.want {
			t.Errorf("formatAmount(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
