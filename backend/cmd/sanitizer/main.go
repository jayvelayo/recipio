package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	_ "modernc.org/sqlite"
)

// Ingredient parser — mirrors the frontend parseIngredient regex.
var (
	amountPat = `(?:[¼½¾⅓⅔⅛⅜⅝⅞]|\d+(?:[./]\d+)?(?:\s+\d+/\d+)?)`
	unitPat   = `(?:cups?|tbsps?|tablespoons?|tsps?|teaspoons?|fl\.?\s*oz|oz|ounces?|lbs?|pounds?|grams?|g|kg|ml|milliliters?|liters?|litres?|l|cloves?|heads?|stalks?|sprigs?|slices?|cans?|packages?|pkgs?|bunches?|handfuls?|pinch(?:es)?|dashes?|pieces?|pcs?)`
	ingRe     = regexp.MustCompile(`(?i)^(` + amountPat + `)\s*(?:(` + unitPat + `)\b[,.]?\s*)?(.*)$`)

	// Strips leading list markers from an ingredient line: "- ", "• ", "* "
	ingPrefixRe = regexp.MustCompile(`^[-•*]+\s*`)

	// Strips leading step numbers: "1. ", "1) ", "Step 1: ", "Step 1. ", etc.
	stepPrefixRe = regexp.MustCompile(`(?i)^(?:step\s*)?\d+[.):\-]\s*`)
)

func parseIngredient(line string) (quantity, name string) {
	line = strings.TrimSpace(ingPrefixRe.ReplaceAllString(strings.TrimSpace(line), ""))
	if line == "" {
		return "", ""
	}
	m := ingRe.FindStringSubmatch(line)
	if m == nil {
		return "", line
	}
	amt := strings.TrimSpace(m[1])
	u := strings.TrimSpace(m[2])
	n := strings.TrimSpace(m[3])
	if n == "" {
		n = line
	}
	parts := []string{}
	if amt != "" {
		parts = append(parts, amt)
	}
	if u != "" {
		parts = append(parts, u)
	}
	return strings.Join(parts, " "), n
}

func sanitizeSteps(encoded string) string {
	steps := strings.Split(encoded, "***")
	for i, s := range steps {
		s = strings.TrimSpace(s)
		s = stepPrefixRe.ReplaceAllString(s, "")
		steps[i] = strings.TrimSpace(s)
	}
	return strings.Join(steps, "***")
}

func main() {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatalf("cannot get cache dir: %v", err)
	}
	dbPath := filepath.Join(cacheDir, "recipio", "recipes.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("cannot open db: %v", err)
	}
	defer db.Close()

	// --- Ingredients ---
	fmt.Println("=== Ingredients ===")
	rows, err := db.Query("SELECT id, name, quantity FROM ingredients")
	if err != nil {
		log.Fatalf("query failed: %v", err)
	}
	type ingRow struct {
		id       int
		name     string
		quantity string
	}
	var ings []ingRow
	for rows.Next() {
		var r ingRow
		if err := rows.Scan(&r.id, &r.name, &r.quantity); err != nil {
			rows.Close()
			log.Fatalf("scan failed: %v", err)
		}
		ings = append(ings, r)
	}
	rows.Close()

	ingStmt, err := db.Prepare("UPDATE ingredients SET name = ?, quantity = ? WHERE id = ?")
	if err != nil {
		log.Fatalf("prepare failed: %v", err)
	}
	defer ingStmt.Close()

	for _, r := range ings {
		original := strings.TrimSpace(r.quantity + " " + r.name)
		newQty, newName := parseIngredient(original)
		fmt.Printf("id=%-3d  %q\n        -> qty=%q  name=%q\n", r.id, original, newQty, newName)
		if _, err := ingStmt.Exec(newName, newQty, r.id); err != nil {
			log.Fatalf("update id=%d failed: %v", r.id, err)
		}
	}

	// --- Steps ---
	fmt.Println("\n=== Steps ===")
	stepRows, err := db.Query("SELECT id, instruction FROM recipes")
	if err != nil {
		log.Fatalf("query failed: %v", err)
	}
	type recipeRow struct {
		id          int
		instruction string
	}
	var recipes []recipeRow
	for stepRows.Next() {
		var r recipeRow
		if err := stepRows.Scan(&r.id, &r.instruction); err != nil {
			stepRows.Close()
			log.Fatalf("scan failed: %v", err)
		}
		recipes = append(recipes, r)
	}
	stepRows.Close()

	stepStmt, err := db.Prepare("UPDATE recipes SET instruction = ? WHERE id = ?")
	if err != nil {
		log.Fatalf("prepare failed: %v", err)
	}
	defer stepStmt.Close()

	for _, r := range recipes {
		sanitized := sanitizeSteps(r.instruction)
		if sanitized == r.instruction {
			continue
		}
		fmt.Printf("id=%-3d  before: %q\n        after:  %q\n", r.id, r.instruction, sanitized)
		if _, err := stepStmt.Exec(sanitized, r.id); err != nil {
			log.Fatalf("update recipe id=%d failed: %v", r.id, err)
		}
	}

	fmt.Printf("\nDone. %d ingredients, %d recipes processed.\n", len(ings), len(recipes))
}
