package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	rec "github.com/jayvelayo/recipio/internal/recipes"
)

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

const testUserID = "test-user-id"

func createFakeServer(db rec.RecipeDatabase) http.Handler {
	mux := http.NewServeMux()
	SetUpRoutes(mux, db, nil, []string{})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), userIDKey, testUserID)
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
}
