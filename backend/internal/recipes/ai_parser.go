package recipes

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

var ErrLLMTimeout = errors.New("llm request timed out")

const defaultLLMAPIURL = "http://localhost:11434/api/generate"

type llmRequest struct {
	Model   string         `json:"model"`
	Prompt  string         `json:"prompt"`
	Format  string         `json:"format"`
	Stream  bool           `json:"stream"`
	Think   bool           `json:"think"`
	Options map[string]any `json:"options"`
}

type llmResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

type AIParser struct {
	APIURL string
	Client *http.Client
}

func NewAIParser() AIParser {
	return AIParser{
		APIURL: defaultLLMAPIURL,
		Client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (p AIParser) ParseRecipeText(text string) (Recipe, error) {
	prompt := fmt.Sprintf(`Extract the recipe into JSON with this schema:
{
  "name": string,
  "description": string,
  "ingredients": [{"name": string, "quantity": string}],
  "instructions": [string]
}

Recipe:
%s

Return ONLY valid JSON.`, text)

	body, err := json.Marshal(llmRequest{
		Model:   "qwen3:4b",
		Prompt:  prompt,
		Format:  "json",
		Stream:  true, // streaming lets Ollama stop mid-generation when the client disconnects (e.g. timeout)
		Think:   false,
		Options: map[string]any{"temperature": 0},
	})
	if err != nil {
		return Recipe{}, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, p.APIURL, bytes.NewReader(body))
	if err != nil {
		return Recipe{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.Client.Do(req)
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return Recipe{}, ErrLLMTimeout
		}
		return Recipe{}, fmt.Errorf("llm request: %w", err)
	}
	defer resp.Body.Close()

	var full strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		var chunk llmResponse
		if err := json.Unmarshal(scanner.Bytes(), &chunk); err != nil {
			return Recipe{}, fmt.Errorf("decode llm chunk: %w", err)
		}
		full.WriteString(chunk.Response)
		if chunk.Done {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return Recipe{}, ErrLLMTimeout
		}
		return Recipe{}, fmt.Errorf("read llm stream: %w", err)
	}

	var recipe Recipe
	if err := json.Unmarshal([]byte(full.String()), &recipe); err != nil {
		return Recipe{}, fmt.Errorf("parse recipe json: %w", err)
	}

	return recipe, nil
}
