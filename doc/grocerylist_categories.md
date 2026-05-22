# Ingredients Category

This doc describes the possible categories of ingredients.

## Categories
- Baking Items
- Bread & Cereals
- Canned Food & Soups
- Dairy, Eggs & Cheese
- Deli
- Meat
- Sauces & Condiments
- Seafood
- Snack & Candy
- Spices & Seasonings
- Fruits
- Vegetable
- Frozen Food

# Architecture

## Schema

Two tables will be created: **ingredient_categories** and **ingredient_category_map**.

The `ingredient_categories` contains all the ingredients that is keyed by integer. This allows faster comparison and prevents duplications.
The `ingredient_category_map` contains all previous mappings found. It contains the normalized ingredient name and references the category_id from `ingredient_categories` table. This allows us to cache AI prompt results.

To promote data isolations, these two related tables can only be accessed through the `CategoryDb` interface. This interface implements:

```go
type CategoryDb interface {
    FetchIngredientCategories(ingredientName []string) []int
    CreateNewMapping(ingredientName string, cat_id int)
}
```

## Normalization

Ingredient names are always normalized to prevent duplicate entries. Each names will be:
- lowercased
- trimmed
- whitespaces converted to '_'

## Evaluation

We will do the lazy evaluation where the ingredients are only parsed and categorized when the grocery list is fetched. This allows us to batch multiple ingredients in a single prompt, thereby saving tokens. The prompt is has a deadline of 5 seconds. Past that, any other uncategorized ingredients remain as default: `Uncategorized`.

## Categorization Function

The categorization is the business logic where it requires both the database and the AI parser. After fetching the grocery list from the `groceryList` module, the list of ingredients will pass-through the categorization function. 

The object will be is implemented as:
```go
// Mutates the []GroceryListItem to recategorize the attach category
func Categorize(items []GroceryListItems, p Prompter, db CategoryDb)
```