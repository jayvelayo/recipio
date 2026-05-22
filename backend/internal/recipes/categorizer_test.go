package recipes

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// fakeCategoryDb is an in-memory implementation of CategoryDb for unit testing
// Categorize / RecategorizeAll without booting SQLite.
type fakeCategoryDb struct {
	mu    sync.Mutex
	cache map[string]int
}

func newFakeCategoryDb() *fakeCategoryDb {
	return &fakeCategoryDb{cache: map[string]int{}}
}

func (f *fakeCategoryDb) FetchIngredientCategories(names []string) ([]int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]int, len(names))
	for i, n := range names {
		out[i] = f.cache[n]
	}
	return out, nil
}

func (f *fakeCategoryDb) CreateNewMapping(name string, catID int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, exists := f.cache[name]; !exists {
		f.cache[name] = catID
	}
	return nil
}

func (f *fakeCategoryDb) AllMappings() (map[string]int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make(map[string]int, len(f.cache))
	for k, v := range f.cache {
		out[k] = v
	}
	return out, nil
}

func (f *fakeCategoryDb) UpdateMapping(name string, catID int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.cache[name] = catID
	return nil
}

// stubPrompter records call counts and replies with a configurable response
// built from a mapping. Only mappings whose key appears in the prompt body
// are echoed back — mimics how the real LLM would only return entries it
// recognized in the input list.
type stubPrompter struct {
	mu       sync.Mutex
	response map[string]string
	calls    int32
}

func newStubPrompter(mapping map[string]string) *stubPrompter {
	return &stubPrompter{response: mapping}
}

func (s *stubPrompter) setResponse(m map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.response = m
}

func (s *stubPrompter) FetchPrompt(_ context.Context, prompt string) (string, error) {
	atomic.AddInt32(&s.calls, 1)
	s.mu.Lock()
	included := map[string]string{}
	for k, v := range s.response {
		if strings.Contains(prompt, k) {
			included[k] = v
		}
	}
	s.mu.Unlock()
	b, _ := json.Marshal(map[string]map[string]string{"categories": included})
	return string(b), nil
}

// blockingPrompter blocks until the caller's context fires its deadline.
// Used to verify the 5s timeout path.
type blockingPrompter struct{}

func (blockingPrompter) FetchPrompt(ctx context.Context, _ string) (string, error) {
	<-ctx.Done()
	return "", ctx.Err()
}

func TestCategorize_CachesViaAI(t *testing.T) {
	prompter := newStubPrompter(map[string]string{
		"chicken_breast": "Meat",
		"flour":          "Baking Items",
	})
	db := newFakeCategoryDb()

	items := []GroceryListItem{
		{Name: "Chicken Breast"},
		{Name: "flour"},
	}
	Categorize(items, prompter, db)

	if atomic.LoadInt32(&prompter.calls) != 1 {
		t.Errorf("expected 1 AI call, got %d", atomic.LoadInt32(&prompter.calls))
	}
	if items[0].Category != "Meat" {
		t.Errorf("chicken breast: want Meat got %q", items[0].Category)
	}
	if items[1].Category != "Baking Items" {
		t.Errorf("flour: want Baking Items got %q", items[1].Category)
	}
	// Cache should now hold both entries.
	all, _ := db.AllMappings()
	if all["chicken_breast"] != CategoryID("Meat") || all["flour"] != CategoryID("Baking Items") {
		t.Errorf("cache did not persist mappings: %v", all)
	}
}

func TestCategorize_UsesCacheOnSecondCall(t *testing.T) {
	prompter := newStubPrompter(map[string]string{"chicken_breast": "Meat"})
	db := newFakeCategoryDb()

	first := []GroceryListItem{{Name: "Chicken Breast"}}
	Categorize(first, prompter, db)
	firstCalls := atomic.LoadInt32(&prompter.calls)

	second := []GroceryListItem{{Name: "Chicken Breast"}}
	Categorize(second, prompter, db)
	if atomic.LoadInt32(&prompter.calls) != firstCalls {
		t.Errorf("expected no additional AI calls after cache warm, got %d additional", atomic.LoadInt32(&prompter.calls)-firstCalls)
	}
	if second[0].Category != "Meat" {
		t.Errorf("expected Meat, got %q", second[0].Category)
	}
}

func TestCategorize_TimeoutFallsBackToUncategorized(t *testing.T) {
	db := newFakeCategoryDb()

	items := []GroceryListItem{{Name: "Mystery Ingredient"}}
	start := time.Now()
	Categorize(items, blockingPrompter{}, db)
	elapsed := time.Since(start)

	if elapsed > 7*time.Second {
		t.Errorf("expected timeout near 5s, took %v", elapsed)
	}
	if items[0].Category != UncategorizedName {
		t.Errorf("expected Uncategorized fallback, got %q", items[0].Category)
	}
}

func TestCategorize_NormalizationProducesCacheHit(t *testing.T) {
	prompter := newStubPrompter(map[string]string{"chicken_breast": "Meat"})
	db := newFakeCategoryDb()

	Categorize([]GroceryListItem{{Name: "Chicken Breast"}}, prompter, db)
	before := atomic.LoadInt32(&prompter.calls)

	variants := []GroceryListItem{{Name: "  chicken   breast  "}}
	Categorize(variants, prompter, db)
	if atomic.LoadInt32(&prompter.calls) != before {
		t.Errorf("normalization should have produced a cache hit, but AI was called again")
	}
	if variants[0].Category != "Meat" {
		t.Errorf("expected Meat, got %q", variants[0].Category)
	}
}

func TestRecategorizeAll_UpdatesCache(t *testing.T) {
	prompter := newStubPrompter(map[string]string{"chicken_breast": "Meat"})
	db := newFakeCategoryDb()

	items := []GroceryListItem{{Name: "Chicken Breast"}}
	Categorize(items, prompter, db)
	if items[0].Category != "Meat" {
		t.Fatalf("expected initial Meat, got %q", items[0].Category)
	}

	// Categories doc shifted: chicken now lives under Deli.
	prompter.setResponse(map[string]string{"chicken_breast": "Deli"})

	total, changed, fellBack, err := RecategorizeAll(prompter, db, 50, 5*time.Second)
	if err != nil {
		t.Fatalf("RecategorizeAll: %v", err)
	}
	if total != 1 || changed != 1 || fellBack != 0 {
		t.Errorf("expected total=1 changed=1 fellBack=0, got %d/%d/%d", total, changed, fellBack)
	}

	after := []GroceryListItem{{Name: "Chicken Breast"}}
	Categorize(after, prompter, db)
	if after[0].Category != "Deli" {
		t.Errorf("expected Deli after recategorize, got %q", after[0].Category)
	}
}
