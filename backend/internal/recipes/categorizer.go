package recipes

import (
	"context"
	"strings"
	"time"
)

// Categories is the ordered list of grocery list categories.
// Mirrors doc/grocerylist_categories.md. Uncategorized is the fallback.
// IDs in the database are 1-indexed positions in this slice — the seeder
// inserts rows with explicit ids so callers can translate between id and
// name without a DB round-trip.
var Categories = []string{
	"Baking Items",
	"Bread & Cereals",
	"Canned Food & Soups",
	"Dairy, Eggs & Cheese",
	"Deli",
	"Meat",
	"Sauces & Condiments",
	"Seafood",
	"Snack & Candy",
	"Spices & Seasonings",
	"Fruits",
	"Vegetable",
	"Frozen Food",
	"Uncategorized",
}

const UncategorizedName = "Uncategorized"

// CategoryID returns the stable id for a category name, or 0 if not known.
func CategoryID(name string) int {
	for i, n := range Categories {
		if n == name {
			return i + 1
		}
	}
	return 0
}

// CategoryName returns the name for a category id, or UncategorizedName for
// out-of-range ids.
func CategoryName(id int) string {
	if id < 1 || id > len(Categories) {
		return UncategorizedName
	}
	return Categories[id-1]
}

const categorizeAITimeout = 5 * time.Second

// CategoryDb is the persistence interface for ingredient categorization.
// Implementations isolate the underlying tables (ingredient_categories,
// ingredient_category_map) — callers never touch them directly.
//
// Names passed in are expected to already be normalized via
// NormalizeIngredientName. Category ids are the stable 1-indexed positions
// from the Categories slice.
type CategoryDb interface {
	// FetchIngredientCategories returns category ids parallel to the input.
	// A returned id of 0 means the name is not in the cache.
	FetchIngredientCategories(names []string) ([]int, error)
	// CreateNewMapping records a new normalized-name → category-id binding.
	// Idempotent: re-creating an existing mapping is not an error.
	CreateNewMapping(name string, catID int) error
	// AllMappings returns every normalized-name → category-id binding in the
	// cache. Used by the recategorize admin tool.
	AllMappings() (map[string]int, error)
	// UpdateMapping replaces the category id for an existing name. Used by
	// the recategorize admin tool when the category list has shifted.
	UpdateMapping(name string, catID int) error
}

// NormalizeIngredientName produces the cache key: lowercase, trimmed, with
// internal whitespace collapsed to single underscores.
func NormalizeIngredientName(name string) string {
	return strings.Join(strings.Fields(strings.ToLower(name)), "_")
}

// Categorize mutates items in place to attach a Category to each. It looks up
// the cache via db, asks the LLM (with a 5s deadline) for anything missing,
// and persists the new mappings. Names left uncategorized after the AI call
// or after a timeout fall back to UncategorizedName — failure to categorize
// never propagates as an error to the caller.
func Categorize(items []GroceryListItem, p Prompter, db CategoryDb) {
	if len(items) == 0 {
		return
	}

	keys := make([]string, len(items))
	var uniq []string
	seen := map[string]bool{}
	for i, item := range items {
		k := NormalizeIngredientName(item.Name)
		keys[i] = k
		if k == "" || seen[k] {
			continue
		}
		seen[k] = true
		uniq = append(uniq, k)
	}

	resolved := map[string]string{}
	var missing []string
	if len(uniq) > 0 {
		ids, err := db.FetchIngredientCategories(uniq)
		if err != nil {
			ids = make([]int, len(uniq))
		}
		for i, k := range uniq {
			if ids[i] > 0 {
				resolved[k] = CategoryName(ids[i])
			} else {
				missing = append(missing, k)
			}
		}
	}

	if len(missing) > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), categorizeAITimeout)
		resp, err := CategorizeIngredients(p, ctx, missing, Categories)
		cancel()
		if err == nil {
			for name, catName := range resp {
				id := CategoryID(catName)
				if id == 0 {
					continue
				}
				if err := db.CreateNewMapping(name, id); err == nil {
					resolved[name] = catName
				}
			}
		}
	}

	for i := range items {
		if cat, ok := resolved[keys[i]]; ok {
			items[i].Category = cat
		} else {
			items[i].Category = UncategorizedName
		}
	}
}

// RecategorizeAll re-runs every cached mapping through the LLM against the
// current Categories list and updates each row. Used by the recategorize CLI
// after the category list has been edited.
func RecategorizeAll(p Prompter, db CategoryDb, batchSize int, batchTimeout time.Duration) (total, changed, fellBack int, err error) {
	all, err := db.AllMappings()
	if err != nil {
		return 0, 0, 0, err
	}
	names := make([]string, 0, len(all))
	for n := range all {
		names = append(names, n)
	}

	uncatID := CategoryID(UncategorizedName)
	for i := 0; i < len(names); i += batchSize {
		end := i + batchSize
		if end > len(names) {
			end = len(names)
		}
		batch := names[i:end]

		ctx, cancel := context.WithTimeout(context.Background(), batchTimeout)
		resp, callErr := CategorizeIngredients(p, ctx, batch, Categories)
		cancel()
		if callErr != nil {
			resp = map[string]string{}
		}

		for _, name := range batch {
			total++
			newCatName, hit := resp[name]
			newCatID := uncatID
			if hit {
				if id := CategoryID(newCatName); id != 0 {
					newCatID = id
				} else {
					newCatName = UncategorizedName
				}
			} else {
				newCatName = UncategorizedName
				fellBack++
			}
			if newCatID != all[name] {
				changed++
			}
			if updErr := db.UpdateMapping(name, newCatID); updErr != nil {
				return total, changed, fellBack, updErr
			}
		}
	}
	return total, changed, fellBack, nil
}
