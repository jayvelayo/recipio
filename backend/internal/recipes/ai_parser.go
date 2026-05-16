package recipes

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
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
  "instructions": [string]
}

Recipe:
%s

Return ONLY valid JSON.`, text)

	body, err := json.Marshal(groqRequest{
		Model:          "llama-3.1-8b-instant",
		Messages:       []groqMessage{{Role: "user", Content: prompt}},
		ResponseFormat: map[string]string{"type": "json_object"},
		Temperature:    0,
	})
	if err != nil {
		return Recipe{}, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, p.APIURL, bytes.NewReader(body))
	if err != nil {
		return Recipe{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	resp, err := p.Client.Do(req)
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return Recipe{}, ErrLLMTimeout
		}
		return Recipe{}, fmt.Errorf("llm request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Recipe{}, fmt.Errorf("groq api error: status %d", resp.StatusCode)
	}

	var groqResp groqResponse
	if err := json.NewDecoder(resp.Body).Decode(&groqResp); err != nil {
		return Recipe{}, fmt.Errorf("decode groq response: %w", err)
	}
	if len(groqResp.Choices) == 0 {
		return Recipe{}, fmt.Errorf("groq returned no choices")
	}

	var recipe Recipe
	if err := json.Unmarshal([]byte(groqResp.Choices[0].Message.Content), &recipe); err != nil {
		return Recipe{}, fmt.Errorf("parse recipe json: %w", err)
	}

	return recipe, nil
}
