package sqlite_db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	rec "github.com/jayvelayo/recipio/internal/recipes"
)

func (iface *SqliteDatabaseContext) CreateRecipe(userID string, newRecipe rec.Recipe) (string, error) {
	db := iface.sqliteDb
	tx, err := db.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	recipeID := uuid.New().String()
	instructions := encodeInstructionList(newRecipe.Instructions)
	_, err = tx.Exec("INSERT INTO recipes (id, user_id, name, instruction, source_url) VALUES (?, ?, ?, ?, ?)", recipeID, userID, newRecipe.Name, instructions, newRecipe.SourceURL)
	if err != nil {
		return "", fmt.Errorf("failed to insert recipe: %w", err)
	}

	ingStmt, err := tx.Prepare("INSERT INTO ingredients (recipe_id, name, quantity) VALUES (?, ?, ?)")
	if err != nil {
		return "", fmt.Errorf("failed to prepare ingredient insert: %w", err)
	}
	defer ingStmt.Close()

	for _, ing := range newRecipe.Ingredients {
		if ing.Name == "" {
			return "", fmt.Errorf("ingredient name cannot be empty")
		}
		if _, err := ingStmt.Exec(recipeID, ing.Name, ing.Quantity); err != nil {
			return "", fmt.Errorf("failed to insert ingredient (%s): %w", ing.Name, err)
		}
	}

	tagStmt, err := tx.Prepare("INSERT OR IGNORE INTO recipe_tags (recipe_id, tag) VALUES (?, ?)")
	if err != nil {
		return "", fmt.Errorf("failed to prepare tag insert: %w", err)
	}
	defer tagStmt.Close()

	for _, tag := range newRecipe.Tags {
		if tag = strings.ToLower(strings.TrimSpace(tag)); tag != "" {
			if _, err := tagStmt.Exec(recipeID, tag); err != nil {
				return "", fmt.Errorf("failed to insert tag (%s): %w", tag, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}
	return recipeID, nil
}

func (ctx *SqliteDatabaseContext) GetRecipe(id string) (rec.Recipe, error) {
	db := ctx.sqliteDb
	var recipe rec.Recipe
	var instructions string

	row := db.QueryRow("SELECT id, name, instruction, COALESCE(source_url, '') FROM recipes WHERE id = ?", id)
	if err := row.Scan(&recipe.ID, &recipe.Name, &instructions, &recipe.SourceURL); err != nil {
		if err == sql.ErrNoRows {
			return rec.Recipe{}, fmt.Errorf("recipe with id %s not found", id)
		}
		return rec.Recipe{}, err
	}
	recipe.Instructions = decodeInstructionList(instructions)

	ingRows, err := db.Query("SELECT name, quantity FROM ingredients WHERE recipe_id = ?", recipe.ID)
	if err != nil {
		return rec.Recipe{}, err
	}
	defer ingRows.Close()

	for ingRows.Next() {
		var ing rec.Ingredient
		if err := ingRows.Scan(&ing.Name, &ing.Quantity); err != nil {
			return rec.Recipe{}, err
		}
		recipe.Ingredients = append(recipe.Ingredients, ing)
	}

	tagRows, err := db.Query("SELECT tag FROM recipe_tags WHERE recipe_id = ? ORDER BY tag", recipe.ID)
	if err != nil {
		return rec.Recipe{}, err
	}
	defer tagRows.Close()

	for tagRows.Next() {
		var tag string
		if err := tagRows.Scan(&tag); err != nil {
			return rec.Recipe{}, err
		}
		recipe.Tags = append(recipe.Tags, tag)
	}
	return recipe, nil
}

func (ctx *SqliteDatabaseContext) UpdateRecipe(id string, recipe rec.Recipe) error {
	db := ctx.sqliteDb
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	instructions := encodeInstructionList(recipe.Instructions)
	res, err := tx.Exec("UPDATE recipes SET name = ?, instruction = ?, source_url = ? WHERE id = ?", recipe.Name, instructions, recipe.SourceURL, id)
	if err != nil {
		return fmt.Errorf("failed to update recipe: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("recipe with id %s not found", id)
	}

	if _, err := tx.Exec("DELETE FROM ingredients WHERE recipe_id = ?", id); err != nil {
		return fmt.Errorf("failed to delete ingredients: %w", err)
	}

	ingStmt, err := tx.Prepare("INSERT INTO ingredients (recipe_id, name, quantity) VALUES (?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare ingredient insert: %w", err)
	}
	defer ingStmt.Close()

	for _, ing := range recipe.Ingredients {
		if ing.Name == "" {
			return fmt.Errorf("ingredient name cannot be empty")
		}
		if _, err := ingStmt.Exec(id, ing.Name, ing.Quantity); err != nil {
			return fmt.Errorf("failed to insert ingredient: %w", err)
		}
	}

	if _, err := tx.Exec("DELETE FROM recipe_tags WHERE recipe_id = ?", id); err != nil {
		return fmt.Errorf("failed to delete tags: %w", err)
	}

	tagStmt, err := tx.Prepare("INSERT OR IGNORE INTO recipe_tags (recipe_id, tag) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare tag insert: %w", err)
	}
	defer tagStmt.Close()

	for _, tag := range recipe.Tags {
		if tag = strings.ToLower(strings.TrimSpace(tag)); tag != "" {
			if _, err := tagStmt.Exec(id, tag); err != nil {
				return fmt.Errorf("failed to insert tag (%s): %w", tag, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (ctx *SqliteDatabaseContext) DeleteRecipe(id string) error {
	db := ctx.sqliteDb
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM ingredients WHERE recipe_id = ?", id); err != nil {
		return fmt.Errorf("failed to delete ingredients: %w", err)
	}
	if _, err := tx.Exec("DELETE FROM recipe_tags WHERE recipe_id = ?", id); err != nil {
		return fmt.Errorf("failed to delete tags: %w", err)
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
		return fmt.Errorf("recipe with id %s not found", id)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (ctx *SqliteDatabaseContext) GetAllRecipes(userID string) (rec.Recipes, error) {
	db := ctx.sqliteDb

	rows, err := db.Query(`
		SELECT r.id, r.name, r.instruction, COALESCE(r.source_url, ''), i.name, i.quantity
		FROM recipes r
		LEFT JOIN ingredients i ON i.recipe_id = r.id
		WHERE r.user_id = ?
		ORDER BY r.id`, userID)
	if err != nil {
		return rec.Recipes{}, err
	}
	defer rows.Close()

	var recipes rec.Recipes
	var current rec.Recipe
	for rows.Next() {
		var id string
		var name, instruction, sourceURL string
		var ingName, ingQty sql.NullString
		if err := rows.Scan(&id, &name, &instruction, &sourceURL, &ingName, &ingQty); err != nil {
			return rec.Recipes{}, err
		}
		if current.ID != id {
			if current.ID != "" {
				recipes = append(recipes, current)
			}
			current = rec.Recipe{
				ID:           id,
				Name:         name,
				Instructions: decodeInstructionList(instruction),
				SourceURL:    sourceURL,
			}
		}
		if ingName.Valid {
			current.Ingredients = append(current.Ingredients, rec.Ingredient{
				Name:     ingName.String,
				Quantity: ingQty.String,
			})
		}
	}
	if current.ID != "" {
		recipes = append(recipes, current)
	}

	if len(recipes) > 0 {
		tagRows, err := db.Query(`
			SELECT rt.recipe_id, rt.tag FROM recipe_tags rt
			INNER JOIN recipes r ON rt.recipe_id = r.id
			WHERE r.user_id = ?
			ORDER BY rt.tag`, userID)
		if err != nil {
			return rec.Recipes{}, err
		}
		defer tagRows.Close()

		tagMap := make(map[string][]string)
		for tagRows.Next() {
			var recipeID, tag string
			if err := tagRows.Scan(&recipeID, &tag); err != nil {
				return rec.Recipes{}, err
			}
			tagMap[recipeID] = append(tagMap[recipeID], tag)
		}
		for i := range recipes {
			if tags, ok := tagMap[recipes[i].ID]; ok {
				recipes[i].Tags = tags
			}
		}
	}

	return recipes, nil
}

func (ctx *SqliteDatabaseContext) AddRecipeToMealPlan(id string) error {
	return nil
}

func (ctx *SqliteDatabaseContext) AddRecipesToMealPlan(mealPlanID string, recipeIDs []string) error {
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

	stmt, err := tx.Prepare("INSERT OR IGNORE INTO meal_plan_recipes (meal_plan_id, recipe_id) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare insert: %w", err)
	}
	defer stmt.Close()

	for _, id := range recipeIDs {
		if id = strings.TrimSpace(id); id == "" {
			continue
		}
		if _, err = stmt.Exec(planID, id); err != nil {
			return fmt.Errorf("failed to add recipe %s: %w", id, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	return nil
}

func (ctx *SqliteDatabaseContext) CreateMealPlan(userID string, recipeIDs []string) (string, error) {
	db := ctx.sqliteDb
	tx, err := db.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.Exec("INSERT INTO meal_plan (user_id) VALUES (?)", userID)
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
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}
		if _, err = stmt.Exec(mealPlanID, idStr); err != nil {
			return "", fmt.Errorf("failed to link recipe %s to meal plan: %w", idStr, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit: %w", err)
	}
	return strconv.FormatInt(mealPlanID, 10), nil
}

func (ctx *SqliteDatabaseContext) GetAllMealPlans(userID string) ([]rec.MealPlanSummary, error) {
	db := ctx.sqliteDb
	planRows, err := db.Query("SELECT id FROM meal_plan WHERE user_id = ? ORDER BY id", userID)
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

func (ctx *SqliteDatabaseContext) GetGroceryList(mealPlanID string) ([]rec.GroceryListItem, error) {
	db := ctx.sqliteDb
	planID, err := strconv.ParseInt(mealPlanID, 10, 64)
	if err != nil || planID < 1 {
		return nil, fmt.Errorf("invalid meal plan id: %s", mealPlanID)
	}

	rows, err := db.Query(
		"SELECT r.id, r.name FROM meal_plan_recipes mpr JOIN recipes r ON r.id = mpr.recipe_id WHERE mpr.meal_plan_id = ?",
		planID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type recipeInfo struct {
		id   string
		name string
	}
	var recipes []recipeInfo
	for rows.Next() {
		var ri recipeInfo
		if err := rows.Scan(&ri.id, &ri.name); err != nil {
			return nil, err
		}
		recipes = append(recipes, ri)
	}

	type group struct {
		name    string
		sources []rec.IngredientSource
	}
	groups := make(map[string]*group)
	var order []string

	for _, r := range recipes {
		ingRows, err := db.Query("SELECT name, quantity FROM ingredients WHERE recipe_id = ?", r.id)
		if err != nil {
			return nil, err
		}
		for ingRows.Next() {
			var name, qty string
			if err := ingRows.Scan(&name, &qty); err != nil {
				ingRows.Close()
				return nil, err
			}
			if name == "" {
				continue
			}
			key := strings.ToLower(strings.TrimSpace(name))
			if _, exists := groups[key]; !exists {
				groups[key] = &group{name: name}
				order = append(order, key)
			}
			groups[key].sources = append(groups[key].sources, rec.IngredientSource{
				Quantity:   qty,
				RecipeName: r.name,
			})
		}
		ingRows.Close()
	}

	result := make([]rec.GroceryListItem, 0, len(order))
	for _, key := range order {
		g := groups[key]
		result = append(result, rec.GroceryListItem{
			Name:    g.name,
			Checked: false,
			Sources: g.sources,
		})
	}
	return result, nil
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

func (ctx *SqliteDatabaseContext) CreateGroceryList(userID string, name string, items []rec.GroceryListItem, mealPlanID *string) (string, error) {
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
		res, err = tx.Exec("INSERT INTO grocery_lists (user_id, name, meal_plan_id) VALUES (?, ?, ?)", userID, name, planID)
		if err != nil {
			return "", fmt.Errorf("failed to insert grocery list: %w", err)
		}
	} else {
		res, err = tx.Exec("INSERT INTO grocery_lists (user_id, name, meal_plan_id) VALUES (?, ?, ?)", userID, name, nil)
		if err != nil {
			return "", fmt.Errorf("failed to insert grocery list: %w", err)
		}
	}

	listID, err := res.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("failed to get grocery list id: %w", err)
	}

	stmt, err := tx.Prepare("INSERT INTO grocery_list_items (grocery_list_id, name, quantity, checked, sources) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return "", fmt.Errorf("failed to prepare item insert: %w", err)
	}
	defer stmt.Close()

	for _, item := range items {
		sourcesJSON, _ := json.Marshal(item.Sources)
		if _, err = stmt.Exec(listID, item.Name, item.Quantity, item.Checked, string(sourcesJSON)); err != nil {
			return "", fmt.Errorf("failed to insert item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit: %w", err)
	}
	return strconv.FormatInt(listID, 10), nil
}

func (ctx *SqliteDatabaseContext) GetAllGroceryLists(userID string) ([]rec.GroceryList, error) {
	db := ctx.sqliteDb
	rows, err := db.Query("SELECT id, name, meal_plan_id FROM grocery_lists WHERE user_id = ? ORDER BY id", userID)
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
	rows, err := db.Query("SELECT name, quantity, checked, sources FROM grocery_list_items WHERE grocery_list_id = ? ORDER BY id", listID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []rec.GroceryListItem
	for rows.Next() {
		var item rec.GroceryListItem
		var sourcesJSON sql.NullString
		if err := rows.Scan(&item.Name, &item.Quantity, &item.Checked, &sourcesJSON); err != nil {
			return nil, err
		}
		if sourcesJSON.Valid && sourcesJSON.String != "" {
			_ = json.Unmarshal([]byte(sourcesJSON.String), &item.Sources)
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

	stmt, err := tx.Prepare("INSERT INTO grocery_list_items (grocery_list_id, name, quantity, checked, sources) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare item insert: %w", err)
	}
	defer stmt.Close()

	for _, item := range items {
		sourcesJSON, _ := json.Marshal(item.Sources)
		if _, err = stmt.Exec(id, item.Name, item.Quantity, item.Checked, string(sourcesJSON)); err != nil {
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
