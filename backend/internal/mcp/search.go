package mcp

import (
	"RealTimeChat/backend/internal/type/model"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func SearchWeb(query string) (string, error) {
	baseURL := "https://api.duckduckgo.com/"
	params := url.Values{}
	params.Set("q", query)
	params.Set("format", "json")
	params.Set("no_html", "1")
	params.Set("skip_disambig", "1")

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// HTTP GET request
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "RealTimeChat/1.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("search API returned status: %d", resp.StatusCode)
	}

	var searchResp model.Search
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("Web search results:\n\n")

	if searchResp.AbstractText != "" {
		sb.WriteString(fmt.Sprintf("Summary: %s\n", searchResp.AbstractText))
		sb.WriteString(fmt.Sprintf("Source: %s\n\n", searchResp.AbstractURL))
	}

	// Related topics
	count := 0
	for _, topic := range searchResp.RelatedTopics {
		if topic.Text == "" || topic.FirstURL == "" {
			continue
		}
		sb.WriteString(fmt.Sprintf("%d. %s\n", count+1, topic.Text))
		sb.WriteString(fmt.Sprintf("   Source: %s\n\n", topic.FirstURL))
		count++
		if count >= 4 {
			break
		}
	}

	if sb.Len() == len("Web search results:\n\n") {
		return "No results found for: " + query, nil
	}

	return sb.String(), nil
}
