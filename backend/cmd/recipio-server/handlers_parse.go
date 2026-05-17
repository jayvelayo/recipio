package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	rec "github.com/jayvelayo/recipio/internal/recipes"
)

func handleDesignParseRecipe(parser rec.AIParser) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}
		var body struct {
			RawRecipeText string `json:"raw_recipe_text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}
		if body.RawRecipeText == "" {
			http.Error(w, "Raw recipe text is required", http.StatusBadRequest)
			return
		}

		recipe, err := parser.ParseRecipeText(sanitizeRecipeText(body.RawRecipeText))
		if err != nil {
			if errors.Is(err, rec.ErrLLMTimeout) {
				http.Error(w, "LLM request timed out", http.StatusGatewayTimeout)
				return
			}
			log.Printf("parse recipe error: %v", err)
			http.Error(w, "Failed to parse recipe", http.StatusInternalServerError)
			return
		}

		encodeJson(w, http.StatusOK, recipe)
	})
}

func handleDesignParseRecipeURL(parser rec.AIParser) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}
		var body struct {
			URL string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}
		if !strings.HasPrefix(body.URL, "http://") && !strings.HasPrefix(body.URL, "https://") {
			http.Error(w, "URL must use http or https", http.StatusBadRequest)
			return
		}
		if len(body.URL) > 2048 {
			http.Error(w, "URL too long", http.StatusBadRequest)
			return
		}

		text, err := fetchAndExtractText(body.URL)
		if err != nil {
			http.Error(w, "Failed to fetch URL: "+err.Error(), http.StatusBadGateway)
			return
		}
		if text == "" {
			http.Error(w, "No text content found at URL", http.StatusUnprocessableEntity)
			return
		}

		recipe, err := parser.ParseRecipeText(text)
		if err != nil {
			if errors.Is(err, rec.ErrLLMTimeout) {
				http.Error(w, "LLM request timed out", http.StatusGatewayTimeout)
				return
			}
			log.Printf("parse recipe url error: %v", err)
			http.Error(w, "Failed to parse recipe", http.StatusInternalServerError)
			return
		}

		encodeJson(w, http.StatusOK, recipe)
	})
}

var (
	reScriptStyle = regexp.MustCompile(`(?i)<(script|style)[^>]*>[\s\S]*?</(script|style)>`)
	reHTMLTag     = regexp.MustCompile(`<[^>]+>`)
	reSpaces      = regexp.MustCompile(`[ \t]+`)
)

func extractTextFromHTML(htmlContent string) string {
	text := reScriptStyle.ReplaceAllString(htmlContent, " ")
	text = reHTMLTag.ReplaceAllString(text, " ")
	text = reSpaces.ReplaceAllString(text, " ")
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	result := strings.Join(lines, "\n")
	if len(result) > 8000 {
		result = result[:8000]
	}
	return result
}

func isPrivateHost(host string) bool {
	if host == "localhost" ||
		strings.HasSuffix(host, ".local") ||
		strings.HasSuffix(host, ".internal") ||
		host == "metadata.google.internal" {
		return true
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified()
}

func fetchAndExtractText(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL")
	}
	if isPrivateHost(parsed.Hostname()) {
		return "", fmt.Errorf("URL not allowed")
	}

	client := &http.Client{
		Timeout: 20 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			if isPrivateHost(req.URL.Hostname()) {
				return fmt.Errorf("redirect to private address not allowed")
			}
			return nil
		},
	}

	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; RecipioBot/1.0; +https://sarap.recipes)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("page returned status %d", resp.StatusCode)
	}

	htmlBytes, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}

	return extractTextFromHTML(string(htmlBytes)), nil
}
