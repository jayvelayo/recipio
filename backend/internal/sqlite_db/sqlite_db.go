package sqlite_db

import (
	"bytes"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"strconv"
	"strings"
	"text/template"

	rec "github.com/jayvelayo/recipio/internal/recipes"
	_ "modernc.org/sqlite"
)

const (
	RecipeTableName          string = "recipes"
	IngredientsTableName     string = "ingredients"
	MealPlanTableName        string = "meal_plan"
	MealPlanRecipesTableName string = "meal_plan_recipes"
)

type SqliteDatabaseContext struct {
	sqliteDb *sql.DB
}

//go:embed schema.tmpl
var tmplFS embed.FS

func getSchema(schemaFile string) (string, error) {
	myTemplate, err := template.ParseFS(tmplFS, schemaFile)
	if err != nil {
		return "", err
	}
	var schemaSQL bytes.Buffer
	if err := myTemplate.Execute(&schemaSQL, nil); err != nil {
		return "", err
	}
	return schemaSQL.String(), nil
}

func InitDb(db_path string) (rec.RecipeDatabase, error) {
	db, err := sql.Open("sqlite", db_path)
	if err != nil {
		log.Fatal(err)
	}
	schema, err := getSchema("schema.tmpl")
	if err != nil {
		return nil, fmt.Errorf("unable to create schema db: %v", err)
	}
	_, err = db.Exec(schema)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize db: %v", err)
	}
	sqliteDb := &SqliteDatabaseContext{
		sqliteDb: db,
	}
	return sqliteDb, nil
}

const encodingChars = "***"

func encodeInstructionList(list rec.InstructionList) string {
	return strings.Join(list, encodingChars)
}

func decodeInstructionList(str string) rec.InstructionList {
	return strings.Split(str, encodingChars)
}

func (iface *SqliteDatabaseContext) CreateRecipe(newRecipe rec.Recipe) (uint64, error) {
	db := iface.sqliteDb
	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert recipe
	instructions := encodeInstructionList(newRecipe.Instructions)
	exec_query := fmt.Sprintf("INSERT INTO %s (name, instruction) VALUES (?, ?)", RecipeTableName)
	res, err := tx.Exec(exec_query, newRecipe.Name, instructions)
	if err != nil {
		return 0, fmt.Errorf("failed to insert recipe: %w\ncmd: exec_query: %s", err, exec_query)
	}

	recipeID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get inserted recipe ID: %w", err)
	}

	// Insert ingredients
	prepare_query := fmt.Sprintf("INSERT INTO %s (recipe_id, name, quantity) VALUES (?, ?, ?)", IngredientsTableName)
	stmt, err := tx.Prepare(prepare_query)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare ingredient insert: %w", err)
	}
	defer stmt.Close()

	for _, ing := range newRecipe.Ingredients {
		if ing.Name == "" {
			return 0, fmt.Errorf("ingredient name cannot be empty")
		}
		_, err := stmt.Exec(recipeID, ing.Name, ing.Quantity)
		if err != nil {
			return 0, fmt.Errorf("failed to insert ingredient (%s): %w", ing.Name, err)
		}
	}
	// Commit if everything is okay
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return uint64(recipeID), nil
}

func (ctx *SqliteDatabaseContext) GetRecipe(id int) (rec.Recipe, error) {
	db := ctx.sqliteDb
	var recipe rec.Recipe
	var instructions string
	// Fetch recipe main data
	query := fmt.Sprintf("SELECT id, name, instruction FROM %s WHERE id = ?", RecipeTableName)
	row := db.QueryRow(query, id)
	if err := row.Scan(&recipe.ID, &recipe.Name, &instructions); err != nil {
		if err == sql.ErrNoRows {
			return rec.Recipe{}, fmt.Errorf("recipe with id %d not found", id)
		}
		return rec.Recipe{}, err
	}
	recipe.Instructions = decodeInstructionList(instructions)

	// Fetch ingredients
	query = fmt.Sprintf("SELECT name, quantity FROM %s WHERE recipe_id = ?", IngredientsTableName)
	rows, err := db.Query(query, recipe.ID)
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
	updateQuery := fmt.Sprintf("UPDATE %s SET name = ?, instruction = ? WHERE id = ?", RecipeTableName)
	res, err := tx.Exec(updateQuery, recipe.Name, instructions, id)
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

	delIngs := fmt.Sprintf("DELETE FROM %s WHERE recipe_id = ?", IngredientsTableName)
	if _, err := tx.Exec(delIngs, id); err != nil {
		return fmt.Errorf("failed to delete ingredients: %w", err)
	}

	stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (recipe_id, name, quantity) VALUES (?, ?, ?)", IngredientsTableName))
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

	// Delete ingredients first (due to FK constraint)
	delIngredients := fmt.Sprintf("DELETE FROM %s WHERE recipe_id = ?", IngredientsTableName)
	if _, err := tx.Exec(delIngredients, id); err != nil {
		return fmt.Errorf("failed to delete ingredients: %w", err)
	}
	// Delete the recipe
	delRecipe := fmt.Sprintf("DELETE FROM %s WHERE id = ?", RecipeTableName)
	res, err := tx.Exec(delRecipe, id)
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
	var recipes rec.Recipes
	db := ctx.sqliteDb

	query := fmt.Sprintf("SELECT id, name, instruction FROM %s", RecipeTableName)
	rows, err := db.Query(query)
	if err != nil {
		return rec.Recipes{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var instructions string
		var recipe rec.Recipe
		if err := rows.Scan(&recipe.ID, &recipe.Name, &instructions); err != nil {
			return recipes, err
		}
		recipe.Instructions = decodeInstructionList(instructions)
		// Fetch ingredients
		query = fmt.Sprintf("SELECT name, quantity FROM %s WHERE recipe_id = ?", IngredientsTableName)
		ing_rows, err := db.Query(query, recipe.ID)
		if err != nil {
			return recipes, err
		}
		for ing_rows.Next() {
			var ing rec.Ingredient
			if err := ing_rows.Scan(&ing.Name, &ing.Quantity); err != nil {
				ing_rows.Close()
				return recipes, err
			}
			recipe.Ingredients = append(recipe.Ingredients, ing)
		}
		ing_rows.Close()
		recipes = append(recipes, recipe)
	}
	return recipes, nil
}

func (ctx *SqliteDatabaseContext) AddRecipeToMealPlan(id int) error {
	return nil
}

const defaultUserID = 0

func (ctx *SqliteDatabaseContext) CreateMealPlan(recipeIDs []string) (string, error) {
	db := ctx.sqliteDb
	tx, err := db.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	insPlan := fmt.Sprintf("INSERT INTO %s (user_id) VALUES (?)", MealPlanTableName)
	res, err := tx.Exec(insPlan, defaultUserID)
	if err != nil {
		return "", fmt.Errorf("failed to insert meal plan: %w", err)
	}
	mealPlanID, err := res.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("failed to get meal plan id: %w", err)
	}

	insLink := fmt.Sprintf("INSERT INTO %s (meal_plan_id, recipe_id) VALUES (?, ?)", MealPlanRecipesTableName)
	stmt, err := tx.Prepare(insLink)
	if err != nil {
		return "", fmt.Errorf("failed to prepare meal_plan_recipes insert: %w", err)
	}
	defer stmt.Close()

	for _, idStr := range recipeIDs {
		id, err := strconv.Atoi(strings.TrimSpace(idStr))
		if err != nil || id < 1 {
			continue
		}
		_, err = stmt.Exec(mealPlanID, id)
		if err != nil {
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
	planQuery := fmt.Sprintf("SELECT id FROM %s ORDER BY id", MealPlanTableName)
	planRows, err := db.Query(planQuery)
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
		namesQuery := fmt.Sprintf(
			"SELECT r.name FROM %s r INNER JOIN %s mpr ON r.id = mpr.recipe_id WHERE mpr.meal_plan_id = ? ORDER BY mpr.recipe_id",
			RecipeTableName, MealPlanRecipesTableName)
		nameRows, err := db.Query(namesQuery, planID)
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

	query := fmt.Sprintf("SELECT recipe_id FROM %s WHERE meal_plan_id = ?", MealPlanRecipesTableName)
	rows, err := db.Query(query, planID)
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
	ingQuery := fmt.Sprintf("SELECT name, quantity FROM %s WHERE recipe_id = ?", IngredientsTableName)
	for _, rid := range recipeIDs {
		ingRows, err := db.Query(ingQuery, rid)
		if err != nil {
			return nil, err
		}
		for ingRows.Next() {
			var name, quantity string
			if err := ingRows.Scan(&name, &quantity); err != nil {
				ingRows.Close()
				return nil, err
			}
			s := strings.TrimSpace(quantity) + " " + strings.TrimSpace(name)
			s = strings.TrimSpace(s)
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

	// Delete from meal_plan_recipes first
	delLinks := fmt.Sprintf("DELETE FROM %s WHERE meal_plan_id = ?", MealPlanRecipesTableName)
	_, err = tx.Exec(delLinks, planID)
	if err != nil {
		return fmt.Errorf("failed to delete meal plan recipes: %w", err)
	}

	// Delete the meal plan
	delPlan := fmt.Sprintf("DELETE FROM %s WHERE id = ?", MealPlanTableName)
	_, err = tx.Exec(delPlan, planID)
	if err != nil {
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

	// Insert grocery list
	insList := "INSERT INTO grocery_lists (name, meal_plan_id) VALUES (?, ?)"
	var res sql.Result
	if mealPlanID != nil {
		planID, err := strconv.ParseInt(*mealPlanID, 10, 64)
		if err != nil {
			return "", fmt.Errorf("invalid meal plan id: %s", *mealPlanID)
		}
		res, err = tx.Exec(insList, name, planID)
	} else {
		res, err = tx.Exec(insList, name, nil)
	}
	if err != nil {
		return "", fmt.Errorf("failed to insert grocery list: %w", err)
	}

	listID, err := res.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("failed to get grocery list id: %w", err)
	}

	// Insert items
	insItem := "INSERT INTO grocery_list_items (grocery_list_id, name, quantity, checked) VALUES (?, ?, ?, ?)"
	stmt, err := tx.Prepare(insItem)
	if err != nil {
		return "", fmt.Errorf("failed to prepare item insert: %w", err)
	}
	defer stmt.Close()

	for _, item := range items {
		_, err = stmt.Exec(listID, item.Name, item.Quantity, item.Checked)
		if err != nil {
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
	query := "SELECT id, name, meal_plan_id FROM grocery_lists ORDER BY id"
	rows, err := db.Query(query)
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

		// Get items
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
	query := "SELECT name, quantity, checked FROM grocery_list_items WHERE grocery_list_id = ? ORDER BY id"
	rows, err := db.Query(query, listID)
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
	query := "SELECT id, name, meal_plan_id FROM grocery_lists WHERE id = ?"
	var list rec.GroceryList
	var mealPlanID sql.NullInt64
	err := db.QueryRow(query, id).Scan(&list.ID, &list.Name, &mealPlanID)
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

	// Delete existing items
	delQuery := "DELETE FROM grocery_list_items WHERE grocery_list_id = ?"
	_, err = tx.Exec(delQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete existing items: %w", err)
	}

	// Insert new items
	insQuery := "INSERT INTO grocery_list_items (grocery_list_id, name, quantity, checked) VALUES (?, ?, ?, ?)"
	stmt, err := tx.Prepare(insQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare item insert: %w", err)
	}
	defer stmt.Close()

	for _, item := range items {
		_, err = stmt.Exec(id, item.Name, item.Quantity, item.Checked)
		if err != nil {
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

	// Delete items first
	delItems := "DELETE FROM grocery_list_items WHERE grocery_list_id = ?"
	_, err = tx.Exec(delItems, id)
	if err != nil {
		return fmt.Errorf("failed to delete grocery list items: %w", err)
	}

	// Delete list
	delList := "DELETE FROM grocery_lists WHERE id = ?"
	_, err = tx.Exec(delList, id)
	if err != nil {
		return fmt.Errorf("failed to delete grocery list: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	return nil
}

func (ctx *SqliteDatabaseContext) CloseDb() {
	ctx.sqliteDb.Close()
}
