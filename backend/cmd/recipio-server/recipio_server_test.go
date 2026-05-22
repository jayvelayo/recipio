package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/jayvelayo/recipio/internal/authn"
	rec "github.com/jayvelayo/recipio/internal/recipes"
)

// noopCategoryDb satisfies rec.CategoryDb with empty results — handler tests
// don't exercise categorization, they assert on status codes only.
type noopCategoryDb struct{}

func (noopCategoryDb) FetchIngredientCategories(names []string) ([]int, error) {
	return make([]int, len(names)), nil
}
func (noopCategoryDb) CreateNewMapping(string, int) error  { return nil }
func (noopCategoryDb) AllMappings() (map[string]int, error) { return map[string]int{}, nil }
func (noopCategoryDb) UpdateMapping(string, int) error      { return nil }

func createRequestWithBody(method, url string, body interface{}) (*http.Request, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func getResponseBody(r *httptest.ResponseRecorder) string {
	bodybytes, err := io.ReadAll(r.Body)
	if err != nil {
		return "unable to read response"
	}
	return string(bodybytes)
}

const (
	testUserID  = "test-user-id"
	otherUserID = "other-user-id"
	asUserHeader = "X-Test-User-Id"
)

func createFakeServer(db rec.RecipeDatabase) http.Handler {
	mux := http.NewServeMux()
	SetUpRoutes(mux, db, noopCategoryDb{}, nil, []string{}, nil, authn.GoogleOAuthConfig{}, authn.EmailSender{})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := testUserID
		if h := r.Header.Get(asUserHeader); h != "" {
			uid = h
		}
		ctx := context.WithValue(r.Context(), userIDKey, uid)
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
}
