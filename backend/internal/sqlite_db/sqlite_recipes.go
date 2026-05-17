package sqlite_db

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	rec "github.com/jayvelayo/recipio/internal/recipes"
)

const defaultUserID = 0

func (iface *SqliteDatabaseContext) CreateRecipe(newRecipe rec.Recipe) (uint64, error) {
	db := iface.sqliteDb
	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	instructions := encodeInstructionList(newRecipe.Instructions)
	res, err := tx.Exec("INSERT INTO recipes (name, instruction) VALUES (?, ?)", newRecipe.Name, instructions)
	if err != nil {
		return 0, fmt.Errorf("failed to insert recipe: %w", err)
	}

	recipeID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get inserted recipe ID: %w", err)
	}

	stmt, err := tx.Prepare("INSERT INTO ingredients (recipe_id, name, quantity) VALUES (?, ?, ?)")
	if err != nil {
		return 0, fmt.Errorf("failed to prepare ingredient insert: %w", err)
	}
	defer stmt.Close()

	for _, ing := range newRecipe.Ingredients {
		if ing.Name == "" {
			return 0, fmt.Errorf("ingredient name cannot be empty")
		}
		if _, err := stmt.Exec(recipeID, ing.Name, ing.Quantity); err != nil {
			return 0, fmt.Errorf("failed to insert ingredient (%s): %w", ing.Name, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return uint64(recipeID), nil
}

func (ctx *SqliteDatabaseContext) GetRecipe(id int) (rec.Recipe, error) {
	db := ctx.sqliteDb
	var recipe rec.Recipe
	var instructions string

	row := db.QueryRow("SELECT id, name, instruction FROM recipes WHERE id = ?", id)
	if err := row.Scan(&recipe.ID, &recipe.Name, &instructions); err != nil {
		if err == sql.ErrNoRows {
			return rec.Recipe{}, fmt.Errorf("recipe with id %d not found", id)
		}
		return rec.Recipe{}, err
	}
	recipe.Instructions = decodeInstructionList(instructions)

	rows, err := db.Query("SELECT name, quantity FROM ingredients WHERE recipe_id = ?", recipe.ID)
	if err != nil {
		return rec.Recipe{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var ing rec.Ingredient
		if err := rows.Scan(&ing.Name, &ing.Quantity); err != nil {
			return rec.Recipe{}, err
		}
		recipe.Ingredients = append(recipe.Ingredients, ing)
	}
	return recipe, nil
}

func (ctx *SqliteDatabaseContext) UpdateRecipe(id int, recipe rec.Recipe) error {
	db := ctx.sqliteDb
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	instructions := encodeInstructionList(recipe.Instructions)
	res, err := tx.Exec("UPDATE recipes SET name = ?, instruction = ? WHERE id = ?", recipe.Name, instructions, id)
	if err != nil {
		return fmt.Errorf("failed to update recipe: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("recipe with id %d not found", id)
	}

	if _, err := tx.Exec("DELETE FROM ingredients WHERE recipe_id = ?", id); err != nil {
		return fmt.Errorf("failed to delete ingredients: %w", err)
	}

	stmt, err := tx.Prepare("INSERT INTO ingredients (recipe_id, name, quantity) VALUES (?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare ingredient insert: %w", err)
	}
	defer stmt.Close()

	for _, ing := range recipe.Ingredients {
		if ing.Name == "" {
			return fmt.Errorf("ingredient name cannot be empty")
		}
		if _, err := stmt.Exec(id, ing.Name, ing.Quantity); err != nil {
			return fmt.Errorf("failed to insert ingredient: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (ctx *SqliteDatabaseContext) DeleteRecipe(id int) error {
	db := ctx.sqliteDb
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM ingredients WHERE recipe_id = ?", id); err != nil {
		return fmt.Errorf("failed to delete ingredients: %w", err)
	}

	res, err := tx.Exec("DELETE FROM recipes WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete recipe: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("recipe with id %d not found", id)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (ctx *SqliteDatabaseContext) GetAllRecipes() (rec.Recipes, error) {
	db := ctx.sqliteDb

	rows, err := db.Query(`
		SELECT r.id, r.name, r.instruction, i.name, i.quantity
		FROM recipes r
		LEFT JOIN ingredients i ON i.recipe_id = r.id
		ORDER BY r.id`)
	if err != nil {
		return rec.Recipes{}, err
	}
	defer rows.Close()

	var recipes rec.Recipes
	var current rec.Recipe
	for rows.Next() {
		var id int
		var name, instruction string
		var ingName, ingQty sql.NullString
		if err := rows.Scan(&id, &name, &instruction, &ingName, &ingQty); err != nil {
			return rec.Recipes{}, err
		}
		if current.ID != id {
			if current.ID != 0 {
				recipes = append(recipes, current)
			}
			current = rec.Recipe{
				ID:           id,
				Name:         name,
				Instructions: decodeInstructionList(instruction),
			}
		}
		if ingName.Valid {
			current.Ingredients = append(current.Ingredients, rec.Ingredient{
				Name:     ingName.String,
				Quantity: ingQty.String,
			})
		}
	}
	if current.ID != 0 {
		recipes = append(recipes, current)
	}
	return recipes, nil
}

func (ctx *SqliteDatabaseContext) AddRecipeToMealPlan(id int) error {
	return nil
}

func (ctx *SqliteDatabaseContext) CreateMealPlan(recipeIDs []string) (string, error) {
	db := ctx.sqliteDb
	tx, err := db.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.Exec("INSERT INTO meal_plan (user_id) VALUES (?)", defaultUserID)
	if err != nil {
		return "", fmt.Errorf("failed to insert meal plan: %w", err)
	}
	mealPlanID, err := res.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("failed to get meal plan id: %w", err)
	}

	stmt, err := tx.Prepare("INSERT INTO meal_plan_recipes (meal_plan_id, recipe_id) VALUES (?, ?)")
	if err != nil {
		return "", fmt.Errorf("failed to prepare meal_plan_recipes insert: %w", err)
	}
	defer stmt.Close()

	for _, idStr := range recipeIDs {
		id, err := strconv.Atoi(strings.TrimSpace(idStr))
		if err != nil || id < 1 {
			continue
		}
		if _, err = stmt.Exec(mealPlanID, id); err != nil {
			return "", fmt.Errorf("failed to link recipe %d to meal plan: %w", id, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit: %w", err)
	}
	return strconv.FormatInt(mealPlanID, 10), nil
}

func (ctx *SqliteDatabaseContext) GetAllMealPlans() ([]rec.MealPlanSummary, error) {
	db := ctx.sqliteDb
	planRows, err := db.Query("SELECT id FROM meal_plan ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer planRows.Close()

	var plans []rec.MealPlanSummary
	for planRows.Next() {
		var planID int64
		if err := planRows.Scan(&planID); err != nil {
			return nil, err
		}
		nameRows, err := db.Query(`
			SELECT r.name FROM recipes r
			INNER JOIN meal_plan_recipes mpr ON r.id = mpr.recipe_id
			WHERE mpr.meal_plan_id = ?
			ORDER BY mpr.recipe_id`, planID)
		if err != nil {
			return nil, err
		}
		var names []string
		for nameRows.Next() {
			var name string
			if err := nameRows.Scan(&name); err != nil {
				nameRows.Close()
				return nil, err
			}
			names = append(names, name)
		}
		nameRows.Close()
		plans = append(plans, rec.MealPlanSummary{
			ID:          strconv.FormatInt(planID, 10),
			RecipeNames: names,
		})
	}
	return plans, nil
}

func (ctx *SqliteDatabaseContext) GetGroceryList(mealPlanID string) ([]string, error) {
	db := ctx.sqliteDb
	planID, err := strconv.ParseInt(mealPlanID, 10, 64)
	if err != nil || planID < 1 {
		return nil, fmt.Errorf("invalid meal plan id: %s", mealPlanID)
	}

	rows, err := db.Query("SELECT recipe_id FROM meal_plan_recipes WHERE meal_plan_id = ?", planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipeIDs []int
	for rows.Next() {
		var rid int
		if err := rows.Scan(&rid); err != nil {
			return nil, err
		}
		recipeIDs = append(recipeIDs, rid)
	}

	var ingredients []string
	for _, rid := range recipeIDs {
		ingRows, err := db.Query("SELECT name, quantity FROM ingredients WHERE recipe_id = ?", rid)
		if err != nil {
			return nil, err
		}
		for ingRows.Next() {
			var name, quantity string
			if err := ingRows.Scan(&name, &quantity); err != nil {
				ingRows.Close()
				return nil, err
			}
			s := strings.TrimSpace(quantity + " " + name)
			if s != "" {
				ingredients = append(ingredients, s)
			}
		}
		ingRows.Close()
	}
	return ingredients, nil
}

func (ctx *SqliteDatabaseContext) DeleteMealPlan(mealPlanID string) error {
	db := ctx.sqliteDb
	planID, err := strconv.ParseInt(mealPlanID, 10, 64)
	if err != nil || planID < 1 {
		return fmt.Errorf("invalid meal plan id: %s", mealPlanID)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err = tx.Exec("DELETE FROM meal_plan_recipes WHERE meal_plan_id = ?", planID); err != nil {
		return fmt.Errorf("failed to delete meal plan recipes: %w", err)
	}
	if _, err = tx.Exec("DELETE FROM meal_plan WHERE id = ?", planID); err != nil {
		return fmt.Errorf("failed to delete meal plan: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	return nil
}

func (ctx *SqliteDatabaseContext) CreateGroceryList(name string, items []rec.GroceryListItem, mealPlanID *string) (string, error) {
	db := ctx.sqliteDb
	tx, err := db.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var res sql.Result
	if mealPlanID != nil {
		planID, err := strconv.ParseInt(*mealPlanID, 10, 64)
		if err != nil {
			return "", fmt.Errorf("invalid meal plan id: %s", *mealPlanID)
		}
		res, err = tx.Exec("INSERT INTO grocery_lists (name, meal_plan_id) VALUES (?, ?)", name, planID)
		if err != nil {
			return "", fmt.Errorf("failed to insert grocery list: %w", err)
		}
	} else {
		res, err = tx.Exec("INSERT INTO grocery_lists (name, meal_plan_id) VALUES (?, ?)", name, nil)
		if err != nil {
			return "", fmt.Errorf("failed to insert grocery list: %w", err)
		}
	}

	listID, err := res.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("failed to get grocery list id: %w", err)
	}

	stmt, err := tx.Prepare("INSERT INTO grocery_list_items (grocery_list_id, name, quantity, checked) VALUES (?, ?, ?, ?)")
	if err != nil {
		return "", fmt.Errorf("failed to prepare item insert: %w", err)
	}
	defer stmt.Close()

	for _, item := range items {
		if _, err = stmt.Exec(listID, item.Name, item.Quantity, item.Checked); err != nil {
			return "", fmt.Errorf("failed to insert item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit: %w", err)
	}
	return strconv.FormatInt(listID, 10), nil
}

func (ctx *SqliteDatabaseContext) GetAllGroceryLists() ([]rec.GroceryList, error) {
	db := ctx.sqliteDb
	rows, err := db.Query("SELECT id, name, meal_plan_id FROM grocery_lists ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []rec.GroceryList
	for rows.Next() {
		var list rec.GroceryList
		var mealPlanID sql.NullInt64
		if err := rows.Scan(&list.ID, &list.Name, &mealPlanID); err != nil {
			return nil, err
		}
		if mealPlanID.Valid {
			mpID := strconv.FormatInt(mealPlanID.Int64, 10)
			list.MealPlanID = &mpID
		}
		items, err := ctx.getGroceryListItems(list.ID)
		if err != nil {
			return nil, err
		}
		list.Items = items
		lists = append(lists, list)
	}
	return lists, nil
}

func (ctx *SqliteDatabaseContext) getGroceryListItems(listID string) ([]rec.GroceryListItem, error) {
	db := ctx.sqliteDb
	rows, err := db.Query("SELECT name, quantity, checked FROM grocery_list_items WHERE grocery_list_id = ? ORDER BY id", listID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []rec.GroceryListItem
	for rows.Next() {
		var item rec.GroceryListItem
		if err := rows.Scan(&item.Name, &item.Quantity, &item.Checked); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (ctx *SqliteDatabaseContext) GetGroceryListByID(id string) (rec.GroceryList, error) {
	db := ctx.sqliteDb
	var list rec.GroceryList
	var mealPlanID sql.NullInt64
	err := db.QueryRow("SELECT id, name, meal_plan_id FROM grocery_lists WHERE id = ?", id).Scan(&list.ID, &list.Name, &mealPlanID)
	if err != nil {
		return rec.GroceryList{}, fmt.Errorf("grocery list not found: %w", err)
	}
	if mealPlanID.Valid {
		mpID := strconv.FormatInt(mealPlanID.Int64, 10)
		list.MealPlanID = &mpID
	}

	items, err := ctx.getGroceryListItems(id)
	if err != nil {
		return rec.GroceryList{}, err
	}
	list.Items = items
	return list, nil
}

func (ctx *SqliteDatabaseContext) UpdateGroceryList(id string, items []rec.GroceryListItem) error {
	db := ctx.sqliteDb
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err = tx.Exec("DELETE FROM grocery_list_items WHERE grocery_list_id = ?", id); err != nil {
		return fmt.Errorf("failed to delete existing items: %w", err)
	}

	stmt, err := tx.Prepare("INSERT INTO grocery_list_items (grocery_list_id, name, quantity, checked) VALUES (?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare item insert: %w", err)
	}
	defer stmt.Close()

	for _, item := range items {
		if _, err = stmt.Exec(id, item.Name, item.Quantity, item.Checked); err != nil {
			return fmt.Errorf("failed to insert item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	return nil
}

func (ctx *SqliteDatabaseContext) DeleteGroceryList(id string) error {
	db := ctx.sqliteDb
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err = tx.Exec("DELETE FROM grocery_list_items WHERE grocery_list_id = ?", id); err != nil {
		return fmt.Errorf("failed to delete grocery list items: %w", err)
	}
	if _, err = tx.Exec("DELETE FROM grocery_lists WHERE id = ?", id); err != nil {
		return fmt.Errorf("failed to delete grocery list: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	return nil
}
