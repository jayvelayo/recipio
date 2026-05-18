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
	// Migrate existing databases that predate the email_verified column.
	_, err = db.Exec(`ALTER TABLE users ADD COLUMN email_verified BOOLEAN NOT NULL DEFAULT FALSE`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		return nil, fmt.Errorf("failed to migrate users table: %w", err)
	}

	// Migrate recipes.id from INTEGER PRIMARY KEY to TEXT PRIMARY KEY so UUIDs can be used.
	var recipesSQL string
	if err := db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='recipes'`).Scan(&recipesSQL); err == nil {
		if strings.Contains(strings.ToUpper(recipesSQL), "ID INTEGER PRIMARY KEY") {
			migrateSteps := []string{
				`CREATE TABLE recipes_new (
					id TEXT PRIMARY KEY,
					user_id TEXT,
					name TEXT NOT NULL,
					instruction TEXT NOT NULL,
					FOREIGN KEY (user_id) REFERENCES users(id)
				)`,
				`INSERT INTO recipes_new (id, user_id, name, instruction)
					SELECT CAST(id AS TEXT), user_id, name, instruction FROM recipes`,
				`DROP TABLE recipes`,
				`ALTER TABLE recipes_new RENAME TO recipes`,
			}
			for _, step := range migrateSteps {
				if _, err := db.Exec(step); err != nil {
					return nil, fmt.Errorf("failed to migrate recipes.id to TEXT: %w", err)
				}
			}
		}
	}
	sqliteDb := &SqliteDatabaseContext{
		sqliteDb: db,
	}
	return sqliteDb, nil
}

func (ctx *SqliteDatabaseContext) CloseDb() {
	ctx.sqliteDb.Close()
}

const encodingChars = "***"

func encodeInstructionList(list rec.InstructionList) string {
	return strings.Join(list, encodingChars)
}

func decodeInstructionList(str string) rec.InstructionList {
	return strings.Split(str, encodingChars)
}
