package sqlite_db_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	rec "github.com/jayvelayo/recipio/internal/recipes"
	"github.com/jayvelayo/recipio/internal/sqlite_db"
)

// fakeGroq returns a test server that responds with the given category map and
// records how many times it was hit (so cache-hit tests can assert no calls were made).
func fakeGroq(t *testing.T, mapping map[string]string) (*httptest.Server, *int32) {
	t.Helper()
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		body, _ := io.ReadAll(r.Body)
		// The prompt embeds each ingredient name on its own line; reply with the
		// configured mapping (ignoring anything not in the map).
		reply := map[string]map[string]string{"categories": {}}
		for k, v := range mapping {
			if strings.Contains(string(body), k) {
				reply["categories"][k] = v
			}
		}
		inner, _ := json.Marshal(reply)
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": string(inner)}},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(srv.Close)
	return srv, &hits
}

func initDBWithParser(t *testing.T, parser rec.AIParser) *sqlite_db.SqliteDatabaseContext {
	t.Helper()
	db, err := sqlite_db.InitDb(":memory:", parser)
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	t.Cleanup(db.CloseDb)
	return db
}

// seedRecipeAndPlan creates a recipe with the given ingredients and a meal plan
// containing it. Returns the meal plan id.
func seedRecipeAndPlan(t *testing.T, db *sqlite_db.SqliteDatabaseContext, ingredients []rec.Ingredient) string {
	t.Helper()
	recipeID, err := db.CreateRecipe("test-user", rec.Recipe{
		Name:         "Test Recipe",
		Instructions: rec.InstructionList{"step"},
		Ingredients:  ingredients,
	})
	if err != nil {
		t.Fatalf("create recipe: %v", err)
	}
	planID, err := db.CreateMealPlan("test-user", []string{recipeID})
	if err != nil {
		t.Fatalf("create meal plan: %v", err)
	}
	return planID
}

func TestGetGroceryList_CategorizesViaAI(t *testing.T) {
	srv, hits := fakeGroq(t, map[string]string{
		"chicken_breast": "Meat",
		"flour":          "Baking Items",
	})
	parser := rec.AIParser{APIURL: srv.URL, Client: &http.Client{Timeout: 5 * time.Second}}
	db := initDBWithParser(t, parser)

	planID := seedRecipeAndPlan(t, db, []rec.Ingredient{
		{Name: "Chicken Breast", Quantity: "1 lb"},
		{Name: "flour", Quantity: "2 cups"},
	})

	items, err := db.GetGroceryList(planID)
	if err != nil {
		t.Fatalf("GetGroceryList: %v", err)
	}
	if atomic.LoadInt32(hits) != 1 {
		t.Errorf("expected exactly 1 AI call, got %d", atomic.LoadInt32(hits))
	}
	gotCat := map[string]string{}
	for _, it := range items {
		gotCat[strings.ToLower(it.Name)] = it.Category
	}
	if gotCat["chicken breast"] != "Meat" {
		t.Errorf("chicken breast category: want Meat got %q", gotCat["chicken breast"])
	}
	if gotCat["flour"] != "Baking Items" {
		t.Errorf("flour category: want Baking Items got %q", gotCat["flour"])
	}
}

func TestGetGroceryList_UsesCacheOnSecondCall(t *testing.T) {
	srv, hits := fakeGroq(t, map[string]string{"chicken_breast": "Meat"})
	parser := rec.AIParser{APIURL: srv.URL, Client: &http.Client{Timeout: 5 * time.Second}}
	db := initDBWithParser(t, parser)

	planID := seedRecipeAndPlan(t, db, []rec.Ingredient{{Name: "Chicken Breast", Quantity: "1 lb"}})

	if _, err := db.GetGroceryList(planID); err != nil {
		t.Fatalf("first call: %v", err)
	}
	firstHits := atomic.LoadInt32(hits)

	items, err := db.GetGroceryList(planID)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if atomic.LoadInt32(hits) != firstHits {
		t.Errorf("expected no additional AI calls after cache warm, got %d additional", atomic.LoadInt32(hits)-firstHits)
	}
	if len(items) != 1 || items[0].Category != "Meat" {
		t.Errorf("expected one item categorized as Meat, got %+v", items)
	}
}

func TestGetGroceryList_TimeoutFallsBackToUncategorized(t *testing.T) {
	// Server that never responds — forces the 5s deadline to fire.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second)
	}))
	t.Cleanup(srv.Close)

	parser := rec.AIParser{APIURL: srv.URL, Client: &http.Client{Timeout: 30 * time.Second}}
	db := initDBWithParser(t, parser)

	planID := seedRecipeAndPlan(t, db, []rec.Ingredient{{Name: "Mystery Ingredient", Quantity: "1"}})

	start := time.Now()
	items, err := db.GetGroceryList(planID)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("GetGroceryList: %v", err)
	}
	if elapsed > 7*time.Second {
		t.Errorf("expected timeout near 5s, took %v", elapsed)
	}
	if len(items) != 1 || items[0].Category != rec.UncategorizedName {
		t.Errorf("expected Uncategorized fallback, got %+v", items)
	}
}

func TestNormalize(t *testing.T) {
	// Indirect check: same normalized key means the second variant should hit
	// the cache populated by the first.
	srv, hits := fakeGroq(t, map[string]string{"chicken_breast": "Meat"})
	parser := rec.AIParser{APIURL: srv.URL, Client: &http.Client{Timeout: 5 * time.Second}}
	db := initDBWithParser(t, parser)

	plan1 := seedRecipeAndPlan(t, db, []rec.Ingredient{{Name: "Chicken Breast"}})
	if _, err := db.GetGroceryList(plan1); err != nil {
		t.Fatalf("first plan: %v", err)
	}

	recipeID2, _ := db.CreateRecipe("test-user", rec.Recipe{
		Name:         "Other",
		Instructions: rec.InstructionList{"step"},
		Ingredients:  []rec.Ingredient{{Name: "  chicken   breast  "}},
	})
	plan2, _ := db.CreateMealPlan("test-user", []string{recipeID2})

	before := atomic.LoadInt32(hits)
	items, err := db.GetGroceryList(plan2)
	if err != nil {
		t.Fatalf("second plan: %v", err)
	}
	if atomic.LoadInt32(hits) != before {
		t.Errorf("normalization should have produced a cache hit, but AI was called again")
	}
	if items[0].Category != "Meat" {
		t.Errorf("expected Meat, got %q", items[0].Category)
	}
}

// mutableMapping wraps the fakeGroq mapping in a place callers can swap atomically.
type mutableMapping struct {
	current atomic.Value // map[string]string
}

func (m *mutableMapping) set(v map[string]string) { m.current.Store(v) }
func (m *mutableMapping) get() map[string]string  { return m.current.Load().(map[string]string) }

func fakeGroqMutable(t *testing.T, initial map[string]string) (*httptest.Server, *mutableMapping) {
	t.Helper()
	mm := &mutableMapping{}
	mm.set(initial)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		reply := map[string]map[string]string{"categories": {}}
		for k, v := range mm.get() {
			if strings.Contains(string(body), k) {
				reply["categories"][k] = v
			}
		}
		inner, _ := json.Marshal(reply)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]string{"content": string(inner)}}},
		})
	}))
	t.Cleanup(srv.Close)
	return srv, mm
}

func TestRecategorizeAll_UpdatesCache(t *testing.T) {
	srv, mm := fakeGroqMutable(t, map[string]string{"chicken_breast": "Meat"})
	parser := rec.AIParser{APIURL: srv.URL, Client: &http.Client{Timeout: 5 * time.Second}}
	db := initDBWithParser(t, parser)

	planID := seedRecipeAndPlan(t, db, []rec.Ingredient{{Name: "Chicken Breast"}})
	items, err := db.GetGroceryList(planID)
	if err != nil {
		t.Fatalf("warm cache: %v", err)
	}
	if items[0].Category != "Meat" {
		t.Fatalf("expected initial Meat, got %q", items[0].Category)
	}

	// Categories doc changed: chicken now lives under Deli.
	mm.set(map[string]string{"chicken_breast": "Deli"})

	total, changed, fellBack, err := db.RecategorizeAll(50, 5*time.Second)
	if err != nil {
		t.Fatalf("RecategorizeAll: %v", err)
	}
	if total != 1 || changed != 1 || fellBack != 0 {
		t.Errorf("expected total=1 changed=1 fellBack=0, got total=%d changed=%d fellBack=%d", total, changed, fellBack)
	}

	items, err = db.GetGroceryList(planID)
	if err != nil {
		t.Fatalf("post-recat list: %v", err)
	}
	if items[0].Category != "Deli" {
		t.Errorf("expected Deli after recategorize, got %q", items[0].Category)
	}
}
