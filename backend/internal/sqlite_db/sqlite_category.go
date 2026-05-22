package sqlite_db

import (
	"fmt"
	"strings"
)

// FetchIngredientCategories returns category ids parallel to names. A returned
// id of 0 means the name is not in ingredient_category_map. Caller is
// responsible for normalizing the names.
func (ctx *SqliteDatabaseContext) FetchIngredientCategories(names []string) ([]int, error) {
	out := make([]int, len(names))
	if len(names) == 0 {
		return out, nil
	}

	placeholders := make([]string, len(names))
	args := make([]any, len(names))
	for i, n := range names {
		placeholders[i] = "?"
		args[i] = n
	}
	q := fmt.Sprintf(
		`SELECT normalized_name, category_id FROM ingredient_category_map WHERE normalized_name IN (%s)`,
		strings.Join(placeholders, ","),
	)
	rows, err := ctx.sqliteDb.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hits := map[string]int{}
	for rows.Next() {
		var name string
		var id int
		if err := rows.Scan(&name, &id); err != nil {
			return nil, err
		}
		hits[name] = id
	}
	for i, n := range names {
		out[i] = hits[n]
	}
	return out, nil
}

func (ctx *SqliteDatabaseContext) CreateNewMapping(name string, catID int) error {
	_, err := ctx.sqliteDb.Exec(
		`INSERT OR IGNORE INTO ingredient_category_map (normalized_name, category_id) VALUES (?, ?)`,
		name, catID,
	)
	return err
}

func (ctx *SqliteDatabaseContext) AllMappings() (map[string]int, error) {
	rows, err := ctx.sqliteDb.Query(`SELECT normalized_name, category_id FROM ingredient_category_map`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var name string
		var id int
		if err := rows.Scan(&name, &id); err != nil {
			return nil, err
		}
		out[name] = id
	}
	return out, nil
}

func (ctx *SqliteDatabaseContext) UpdateMapping(name string, catID int) error {
	_, err := ctx.sqliteDb.Exec(
		`UPDATE ingredient_category_map SET category_id = ? WHERE normalized_name = ?`,
		catID, name,
	)
	return err
}
