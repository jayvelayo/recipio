// recategorize re-runs every cached ingredient through the AI categorizer
// against the current ingredient_categories table. Run this after editing
// doc/grocerylist_categories.md (and the matching list in
// internal/recipes/categories.go) so existing cache rows are reassigned to the
// new category set.
//
// Usage (from the backend/ directory):
//
//	GROQ_API_KEY=... RECIPIO_DB_PATH=... go run ./cmd/recategorize
package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

	rec "github.com/jayvelayo/recipio/internal/recipes"
	"github.com/jayvelayo/recipio/internal/sqlite_db"
)

func main() {
	batchSize := flag.Int("batch", 50, "ingredients per AI call")
	batchTimeout := flag.Duration("timeout", 30*time.Second, "per-batch AI timeout")
	flag.Parse()

	dbPath := os.Getenv("RECIPIO_DB_PATH")
	if dbPath == "" {
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			log.Fatalf("unable to get user cache dir: %v", err)
		}
		dbPath = filepath.Join(cacheDir, "recipio", "recipes.db")
	}
	log.Printf("using database: %s", dbPath)

	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		log.Fatal("GROQ_API_KEY must be set")
	}
	parser := rec.NewAIParser(apiKey)

	db, err := sqlite_db.InitDb(dbPath, parser)
	if err != nil {
		log.Fatalf("init db: %v", err)
	}
	defer db.CloseDb()

	total, changed, fellBack, err := db.RecategorizeAll(*batchSize, *batchTimeout)
	if err != nil {
		log.Fatalf("recategorize: %v", err)
	}
	log.Printf("recategorize done: processed=%d changed=%d uncategorized=%d", total, changed, fellBack)
}
