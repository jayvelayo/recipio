package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	_ "modernc.org/sqlite"
)

type SqliteDatabaseContext struct {
	sqliteDb *sql.DB
	schema   SchemaData
}

type SchemaData struct {
	RecipesTable     string
	IngredientsTable string
}

func applySchema(schemaFile string, data SchemaData) (string, error) {
	tmplContent, err := os.ReadFile(schemaFile)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("schema").Parse(string(tmplContent))
	if err != nil {
		return "", err
	}

	var schemaSQL bytes.Buffer
	if err := tmpl.Execute(&schemaSQL, data); err != nil {
		return "", err
	}
	return schemaSQL.String(), nil
}

func initDb() (RecipeDatabase, error) {
	db, err := sql.Open("sqlite", "recipes.db")
	if err != nil {
		log.Fatal(err)
	}
	schemaData := SchemaData{
		RecipesTable:     "recipes",
		IngredientsTable: "ingredients",
	}
	schema, err := applySchema("./schema.tmpl", schemaData)
	if err != nil {
		return nil, fmt.Errorf("unable to create schema db: %v", err)
	}
	_, err = db.Exec(schema)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize db: %v", err)
	}
	sqliteDb := &SqliteDatabaseContext{
		sqliteDb: db,
		schema:   schemaData,
	}
	return sqliteDb, nil
}

const encodingChars = "***"

func encodeInstructionList(list instructionList) string {
	return strings.Join(list, encodingChars)
}

func decodeInstructionList(str string) instructionList {
	return strings.Split(str, encodingChars)
}

func (iface *SqliteDatabaseContext) createRecipe(newRecipe Recipe) (uint64, error) {
	db := iface.sqliteDb
	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert recipe
	instructions := encodeInstructionList(newRecipe.Instructions)
	exec_query := fmt.Sprintf("INSERT INTO %s (name, instruction) VALUES (?, ?)", iface.schema.RecipesTable)
	res, err := tx.Exec(exec_query, newRecipe.Name, instructions)
	if err != nil {
		return 0, fmt.Errorf("failed to insert recipe: %w\ncmd: exec_query: %s", err, exec_query)
	}

	recipeID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get inserted recipe ID: %w", err)
	}

	// Insert ingredients
	prepare_query := fmt.Sprintf("INSERT INTO %s (recipe_id, name, quantity) VALUES (?, ?, ?)", iface.schema.IngredientsTable)
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

func (ctx *SqliteDatabaseContext) getRecipe(id int) (Recipe, error) {
	db := ctx.sqliteDb
	var recipe Recipe
	var instructions string
	// Fetch recipe main data
	query := fmt.Sprintf("SELECT id, name, instruction FROM %s WHERE id = ?", ctx.schema.RecipesTable)
	row := db.QueryRow(query, id)
	if err := row.Scan(&recipe.ID, &recipe.Name, &instructions); err != nil {
		if err == sql.ErrNoRows {
			return Recipe{}, fmt.Errorf("recipe with id %d not found", id)
		}
		return Recipe{}, err
	}
	recipe.Instructions = decodeInstructionList(instructions)

	// Fetch ingredients
	query = fmt.Sprintf("SELECT name, quantity FROM %s WHERE recipe_id = ?", ctx.schema.IngredientsTable)
	rows, err := db.Query(query, recipe.ID)
	if err != nil {
		return Recipe{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var ing Ingredient
		if err := rows.Scan(&ing.Name, &ing.Quantity); err != nil {
			return Recipe{}, err
		}
		recipe.Ingredients = append(recipe.Ingredients, ing)
	}
	return recipe, nil
}

func (ctx *SqliteDatabaseContext) getAllRecipes() (Recipes, error) {
	var recipes Recipes
	db := ctx.sqliteDb

	query := fmt.Sprintf("SELECT id, name, instruction FROM %s", ctx.schema.RecipesTable)
	rows, err := db.Query(query)
	if err != nil {
		return Recipes{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var instructions string
		var recipe Recipe
		if err := rows.Scan(&recipe.ID, &recipe.Name, &instructions); err != nil {
			return recipes, err
		}
		recipe.Instructions = decodeInstructionList(instructions)
		// Fetch ingredients
		query = fmt.Sprintf("SELECT name, quantity FROM %s WHERE recipe_id = ?", ctx.schema.IngredientsTable)
		ing_rows, err := db.Query(query, recipe.ID)
		if err != nil {
			return recipes, err
		}
		defer rows.Close()
		log.Println("Found recipe id: %d", recipe.ID)
		for ing_rows.Next() {
			var ing Ingredient
			if err := ing_rows.Scan(&ing.Name, &ing.Quantity); err != nil {
				return recipes, err
			}
			recipe.Ingredients = append(recipe.Ingredients, ing)
		}
		recipes = append(recipes, recipe)
	}
	return recipes, nil
}

func (ctx *SqliteDatabaseContext) closeDb() {
	ctx.sqliteDb.Close()
}
