package sqlite_db

import (
	"bytes"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"strings"
	"text/template"

	rec "github.com/jayvelayo/recipio/internal/recipes"
	_ "modernc.org/sqlite"
)

const (
	RecipeTableName      string = "recipes"
	IngredientsTableName string = "ingredients"
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
		defer rows.Close()
		log.Printf("Found recipe id: %d", recipe.ID)
		for ing_rows.Next() {
			var ing rec.Ingredient
			if err := ing_rows.Scan(&ing.Name, &ing.Quantity); err != nil {
				return recipes, err
			}
			recipe.Ingredients = append(recipe.Ingredients, ing)
		}
		recipes = append(recipes, recipe)
	}
	return recipes, nil
}

func (ctx *SqliteDatabaseContext) AddRecipeToMealPlan(id int) error {
	return nil
}

func (ctx *SqliteDatabaseContext) CloseDb() {
	ctx.sqliteDb.Close()
}
