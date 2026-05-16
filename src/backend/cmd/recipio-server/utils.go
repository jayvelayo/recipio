package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"unicode"
)

func decodeJson[T any](r *http.Request) (T, error) {
	var v T
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&v)
	if err != nil {
		return v, fmt.Errorf("json decode error %w", err)
	}
	return v, err
}

func encodeJson[T any](w http.ResponseWriter, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(v)
	return nil
}

func sanitizeRecipeText(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			return r
		}
		switch r {
		case '.', ',', '-', '/', '(', ')', '°', '\'', '"', ':':
			return r
		}
		return -1
	}, s)
}

func returnError(w http.ResponseWriter, httpStatus int, status ResponseStatus, errorMessage string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	errReponse := CreateRecipeResponse{
		Status:       status,
		ErrorMessage: errorMessage,
		RecipeId:     0,
	}
	json.NewEncoder(w).Encode(errReponse)
}
