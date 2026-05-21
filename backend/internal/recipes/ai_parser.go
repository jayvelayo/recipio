package recipes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

var ErrLLMTimeout = errors.New("llm request timed out")

const groqAPIURL = "https://api.groq.com/openai/v1/chat/completions"

type AIParser struct {
	APIURL string
	APIKey string
	Client *http.Client
}

func NewAIParser(apiKey string) AIParser {
	return AIParser{
		APIURL: groqAPIURL,
		APIKey: apiKey,
		Client: &http.Client{Timeout: 60 * time.Second},
	}
}

// callGroq sends one chat-completion request and returns the message content.
// Passes ctx through so callers can enforce per-call deadlines tighter than the client timeout.
func (p AIParser) callGroq(ctx context.Context, prompt string) (string, error) {
	body, err := json.Marshal(groqRequest{
		Model:          "llama-3.1-8b-instant",
		Messages:       []groqMessage{{Role: "user", Content: prompt}},
		ResponseFormat: map[string]string{"type": "json_object"},
		Temperature:    0,
	})
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.APIURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	resp, err := p.Client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return "", ErrLLMTimeout
		}
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return "", ErrLLMTimeout
		}
		return "", fmt.Errorf("llm request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("groq api error: status %d", resp.StatusCode)
	}

	var groqResp groqResponse
	if err := json.NewDecoder(resp.Body).Decode(&groqResp); err != nil {
		return "", fmt.Errorf("decode groq response: %w", err)
	}
	if len(groqResp.Choices) == 0 {
		return "", fmt.Errorf("groq returned no choices")
	}
	return groqResp.Choices[0].Message.Content, nil
}

type groqRequest struct {
	Model          string            `json:"model"`
	Messages       []groqMessage     `json:"messages"`
	ResponseFormat map[string]string `json:"response_format"`
	Temperature    float64           `json:"temperature"`
}

type groqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (p AIParser) ParseRecipeText(text string) (Recipe, error) {
	prompt := fmt.Sprintf(`Extract the recipe into JSON with this schema:
{
  "name": string,
  "description": string,
  "ingredients": [{"name": string, "quantity": string}],
  "instructions": [string],
  "tags": [string]
}

For "tags", infer 1-3 short lowercase tags describing the cuisine or diet
(e.g. "italian", "vegetarian", "thai", "gluten-free"). Omit if unclear.

Recipe:
%s

Return ONLY valid JSON.`, text)

	content, err := p.callGroq(context.Background(), prompt)
	if err != nil {
		return Recipe{}, err
	}
	var recipe Recipe
	if err := json.Unmarshal([]byte(content), &recipe); err != nil {
		return Recipe{}, fmt.Errorf("parse recipe json: %w", err)
	}
	return recipe, nil
}

// CategorizeIngredients asks the model to place each ingredient name into one of the
// allowed categories. Returns a map keyed by the input names (caller decides the
// normalization scheme). Names the model can't place are simply absent from the result;
// the caller is responsible for falling back to Uncategorized.
func (p AIParser) CategorizeIngredients(ctx context.Context, names []string, categories []string) (map[string]string, error) {
	if len(names) == 0 {
		return map[string]string{}, nil
	}
	allowed := strings.Join(categories, ", ")
	listed := strings.Join(names, "\n")
	prompt := fmt.Sprintf(`Categorize each grocery ingredient into exactly one of these categories:
%s

Return JSON with the schema:
{ "categories": { "<ingredient-name>": "<category>", ... } }

Use the ingredient names exactly as given. Pick the single best category from the
list above for each ingredient. If none fit, use "%s".

Ingredients:
%s

Return ONLY valid JSON.`, allowed, UncategorizedName, listed)

	content, err := p.callGroq(ctx, prompt)
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Categories map[string]string `json:"categories"`
	}
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return nil, fmt.Errorf("parse categorize response: %w", err)
	}
	valid := make(map[string]struct{}, len(categories))
	for _, c := range categories {
		valid[c] = struct{}{}
	}
	result := make(map[string]string, len(parsed.Categories))
	for name, cat := range parsed.Categories {
		if _, ok := valid[cat]; ok {
			result[name] = cat
		}
	}
	return result, nil
}
