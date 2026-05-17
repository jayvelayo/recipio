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
