package sqlite_db

import (
	"context"
	"fmt"
	"strings"
	"time"

	rec "github.com/jayvelayo/recipio/internal/recipes"
)

const categorizeAITimeout = 5 * time.Second

// normalizeIngredientName produces the cache key used in ingredient_category_cache:
// lowercase, trimmed, with internal whitespace collapsed to single underscores.
func normalizeIngredientName(name string) string {
	return strings.Join(strings.Fields(strings.ToLower(name)), "_")
}

// attachCategories fills the Category field on each item by looking up the cache.
// Items whose normalized name is not in the cache are sent in a single AI batch with
// a 5s deadline. Anything still uncategorized after that falls back to Uncategorized.
// Never returns an error to the caller's flow — categorization failures degrade
// to Uncategorized rather than breaking grocery list reads.
func (ctx *SqliteDatabaseContext) attachCategories(items []rec.GroceryListItem) []rec.GroceryListItem {
	if len(items) == 0 {
		return items
	}

	keys := make([]string, len(items))
	uniq := make(map[string]struct{}, len(items))
	for i, item := range items {
		k := normalizeIngredientName(item.Name)
		keys[i] = k
		if k != "" {
			uniq[k] = struct{}{}
		}
	}

	cached, err := ctx.lookupCachedCategories(uniq)
	if err != nil {
		cached = map[string]string{}
	}

	var missing []string
	for k := range uniq {
		if _, hit := cached[k]; !hit {
			missing = append(missing, k)
		}
	}

	if len(missing) > 0 {
		fresh := ctx.categorizeAndCache(missing)
		for k, v := range fresh {
			cached[k] = v
		}
	}

	for i := range items {
		if cat, ok := cached[keys[i]]; ok {
			items[i].Category = cat
		} else {
			items[i].Category = rec.UncategorizedName
		}
	}
	return items
}

func (ctx *SqliteDatabaseContext) lookupCachedCategories(keys map[string]struct{}) (map[string]string, error) {
	if len(keys) == 0 {
		return map[string]string{}, nil
	}
	placeholders := make([]string, 0, len(keys))
	args := make([]interface{}, 0, len(keys))
	for k := range keys {
		placeholders = append(placeholders, "?")
		args = append(args, k)
	}
	q := fmt.Sprintf(
		`SELECT c.normalized_name, cat.name
		 FROM ingredient_category_cache c
		 JOIN ingredient_categories cat ON cat.id = c.category_id
		 WHERE c.normalized_name IN (%s)`,
		strings.Join(placeholders, ","),
	)
	rows, err := ctx.sqliteDb.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string, len(keys))
	for rows.Next() {
		var key, cat string
		if err := rows.Scan(&key, &cat); err != nil {
			return nil, err
		}
		result[key] = cat
	}
	return result, nil
}

// categorizeAndCache asks the AI to categorize the missing keys, writes the
// results to the cache, and returns them. On AI failure / timeout, returns an
// empty map — callers fall back to Uncategorized.
func (ctx *SqliteDatabaseContext) categorizeAndCache(missing []string) map[string]string {
	aiCtx, cancel := context.WithTimeout(context.Background(), categorizeAITimeout)
	defer cancel()

	resp, err := ctx.parser.CategorizeIngredients(aiCtx, missing, rec.Categories)
	if err != nil || len(resp) == 0 {
		return map[string]string{}
	}

	catIDs, err := ctx.categoryNameToID()
	if err != nil {
		return map[string]string{}
	}

	stmt, err := ctx.sqliteDb.Prepare(
		`INSERT OR IGNORE INTO ingredient_category_cache (normalized_name, category_id) VALUES (?, ?)`,
	)
	if err != nil {
		return map[string]string{}
	}
	defer stmt.Close()

	out := make(map[string]string, len(resp))
	for name, cat := range resp {
		id, ok := catIDs[cat]
		if !ok {
			continue
		}
		if _, err := stmt.Exec(name, id); err != nil {
			continue
		}
		out[name] = cat
	}
	return out
}

func (ctx *SqliteDatabaseContext) categoryNameToID() (map[string]int64, error) {
	rows, err := ctx.sqliteDb.Query(`SELECT id, name FROM ingredient_categories`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[string]int64{}
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		result[name] = id
	}
	return result, nil
}

// recategorizeAll is used by the recategorize CLI. It re-runs every row in
// ingredient_category_cache through the AI parser against the current category set,
// updates each row's category_id, and returns counts (total, changed, fellBack).
// Unlike attachCategories, this uses a longer per-batch deadline because it's an
// offline job, not a request-path call.
func (ctx *SqliteDatabaseContext) RecategorizeAll(batchSize int, batchTimeout time.Duration) (total, changed, fellBack int, err error) {
	rows, err := ctx.sqliteDb.Query(
		`SELECT c.normalized_name, cat.name
		 FROM ingredient_category_cache c
		 JOIN ingredient_categories cat ON cat.id = c.category_id`,
	)
	if err != nil {
		return 0, 0, 0, err
	}
	type entry struct{ name, oldCat string }
	var all []entry
	for rows.Next() {
		var e entry
		if err := rows.Scan(&e.name, &e.oldCat); err != nil {
			rows.Close()
			return 0, 0, 0, err
		}
		all = append(all, e)
	}
	rows.Close()

	catIDs, err := ctx.categoryNameToID()
	if err != nil {
		return 0, 0, 0, err
	}
	categories := make([]string, 0, len(catIDs))
	for name := range catIDs {
		categories = append(categories, name)
	}

	uncatID, ok := catIDs[rec.UncategorizedName]
	if !ok {
		return 0, 0, 0, fmt.Errorf("Uncategorized category missing from ingredient_categories table")
	}

	updateStmt, err := ctx.sqliteDb.Prepare(`UPDATE ingredient_category_cache SET category_id = ? WHERE normalized_name = ?`)
	if err != nil {
		return 0, 0, 0, err
	}
	defer updateStmt.Close()

	for i := 0; i < len(all); i += batchSize {
		end := i + batchSize
		if end > len(all) {
			end = len(all)
		}
		batch := all[i:end]
		names := make([]string, len(batch))
		for j, e := range batch {
			names[j] = e.name
		}

		aiCtx, cancel := context.WithTimeout(context.Background(), batchTimeout)
		resp, callErr := ctx.parser.CategorizeIngredients(aiCtx, names, categories)
		cancel()
		if callErr != nil {
			resp = map[string]string{}
		}

		for _, e := range batch {
			total++
			newCat, hit := resp[e.name]
			newCatID := uncatID
			if hit {
				if id, ok := catIDs[newCat]; ok {
					newCatID = id
				} else {
					newCat = rec.UncategorizedName
				}
			} else {
				newCat = rec.UncategorizedName
				fellBack++
			}
			if newCat != e.oldCat {
				changed++
			}
			if _, err := updateStmt.Exec(newCatID, e.name); err != nil {
				return total, changed, fellBack, err
			}
		}
	}
	return total, changed, fellBack, nil
}
