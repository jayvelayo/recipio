package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

const groqAPIURL = "https://api.groq.com/openai/v1/chat/completions"

type GroqClient struct {
	APIURL string
	APIKey string
	Client *http.Client
}

func NewGroqClient(apiKey string) GroqClient {
	return GroqClient{
		APIURL: groqAPIURL,
		APIKey: apiKey,
		Client: &http.Client{Timeout: 60 * time.Second},
	}
}

// FetchPrompt sends one chat-completion request and returns the message content.
// Passes ctx through so callers can enforce per-call deadlines tighter than the client timeout.
func (c GroqClient) FetchPrompt(ctx context.Context, prompt string) (string, error) {
	body, err := json.Marshal(groqRequest{
		Model:          "llama-3.1-8b-instant",
		Messages:       []groqMessage{{Role: "user", Content: prompt}},
		ResponseFormat: map[string]string{"type": "json_object"},
		Temperature:    0,
	})
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.APIURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.Client.Do(req)
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

	var parsed groqResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("decode groq response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("groq returned no choices")
	}
	return parsed.Choices[0].Message.Content, nil
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
