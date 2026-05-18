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

type migration struct {
	version int
	apply   func(*sql.DB) error
}

var dbMigrations = []migration{
	{1, func(db *sql.DB) error {
		_, err := db.Exec(`ALTER TABLE users ADD COLUMN email_verified BOOLEAN NOT NULL DEFAULT FALSE`)
		if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
			return err
		}
		return nil
	}},
	{2, func(db *sql.DB) error {
		var recipesSQL string
		if err := db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='recipes'`).Scan(&recipesSQL); err != nil {
			return err
		}
		if !strings.Contains(strings.ToUpper(recipesSQL), "ID INTEGER PRIMARY KEY") {
			return nil
		}
		for _, step := range []string{
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
		} {
			if _, err := db.Exec(step); err != nil {
				return err
			}
		}
		return nil
	}},
	{3, func(db *sql.DB) error {
		_, err := db.Exec(`ALTER TABLE grocery_list_items ADD COLUMN sources TEXT`)
		if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
			return err
		}
		return nil
	}},
	{4, func(db *sql.DB) error {
		_, err := db.Exec(`ALTER TABLE recipes ADD COLUMN source_url TEXT`)
		if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
			return err
		}
		return nil
	}},
	{5, func(db *sql.DB) error {
		_, err := db.Exec(`ALTER TABLE recipes ADD COLUMN is_favourite BOOLEAN NOT NULL DEFAULT FALSE`)
		if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
			return err
		}
		return nil
	}},
}

func runMigrations(db *sql.DB) error {
	var current int
	if err := db.QueryRow("PRAGMA user_version").Scan(&current); err != nil {
		return fmt.Errorf("failed to read schema version: %w", err)
	}
	for _, m := range dbMigrations {
		if current >= m.version {
			continue
		}
		if err := m.apply(db); err != nil {
			return fmt.Errorf("migration to v%d failed: %w", m.version, err)
		}
		if _, err := db.Exec(fmt.Sprintf("PRAGMA user_version = %d", m.version)); err != nil {
			return fmt.Errorf("failed to set schema version to %d: %w", m.version, err)
		}
		current = m.version
	}
	return nil
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
	if _, err = db.Exec(schema); err != nil {
		return nil, fmt.Errorf("unable to initialize db: %v", err)
	}
	if err := runMigrations(db); err != nil {
		return nil, err
	}
	return &SqliteDatabaseContext{sqliteDb: db}, nil
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
