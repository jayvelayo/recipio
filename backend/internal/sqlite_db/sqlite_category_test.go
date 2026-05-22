package sqlite_db_test

import (
	"testing"

	rec "github.com/jayvelayo/recipio/internal/recipes"
	"github.com/jayvelayo/recipio/internal/sqlite_db"
)

func initCategoryDB(t *testing.T) *sqlite_db.SqliteDatabaseContext {
	t.Helper()
	db, err := sqlite_db.InitDb(":memory:")
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	t.Cleanup(db.CloseDb)
	return db
}

func TestCategoryDb_FetchAndCreate(t *testing.T) {
	db := initCategoryDB(t)

	// All misses on an empty cache.
	ids, err := db.FetchIngredientCategories([]string{"chicken_breast", "flour"})
	if err != nil {
		t.Fatalf("fetch empty: %v", err)
	}
	if ids[0] != 0 || ids[1] != 0 {
		t.Errorf("expected zeros for misses, got %v", ids)
	}

	meatID := rec.CategoryID("Meat")
	bakeID := rec.CategoryID("Baking Items")
	if err := db.CreateNewMapping("chicken_breast", meatID); err != nil {
		t.Fatalf("create chicken: %v", err)
	}
	if err := db.CreateNewMapping("flour", bakeID); err != nil {
		t.Fatalf("create flour: %v", err)
	}

	ids, err = db.FetchIngredientCategories([]string{"chicken_breast", "unknown", "flour"})
	if err != nil {
		t.Fatalf("fetch after create: %v", err)
	}
	if ids[0] != meatID || ids[1] != 0 || ids[2] != bakeID {
		t.Errorf("expected [%d 0 %d], got %v", meatID, bakeID, ids)
	}
}

func TestCategoryDb_AllAndUpdate(t *testing.T) {
	db := initCategoryDB(t)
	meatID := rec.CategoryID("Meat")
	deliID := rec.CategoryID("Deli")

	_ = db.CreateNewMapping("chicken_breast", meatID)
	_ = db.CreateNewMapping("ham", meatID)

	all, err := db.AllMappings()
	if err != nil {
		t.Fatalf("all: %v", err)
	}
	if all["chicken_breast"] != meatID || all["ham"] != meatID {
		t.Errorf("unexpected mappings: %v", all)
	}

	if err := db.UpdateMapping("ham", deliID); err != nil {
		t.Fatalf("update: %v", err)
	}
	all, _ = db.AllMappings()
	if all["ham"] != deliID {
		t.Errorf("update did not stick: %v", all)
	}
}
