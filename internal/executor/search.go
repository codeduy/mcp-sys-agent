package executor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type SearxResult struct {
	Results []struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Content string `json:"content"`
	} `json:"results"`
}

func SearchInternet(query string) (string, error) {
	// Replace with Tailscale IP that running SearxNG
	const gatewayIP = "100.108.61.12"

	searchURL := fmt.Sprintf("http://%s:8080/search?q=%s&format=json", gatewayIP, url.QueryEscape(query))

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(searchURL)
	if err != nil {
		return "", fmt.Errorf("Cannot connect to SearxNG: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("SearxNG returned non-OK status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read SearxNG response body: %v", err)
	}
	var data SearxResult
	if err := json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("Error occur while parse data: %v", err)
	}

	var output string
	for i, r := range data.Results {
		if i >= 5 { break } // Get top 5 results
		output += fmt.Sprintf("[%d] %s\nURL: %s\nSnippet: %s\n\n", i+1, r.Title, r.URL, r.Content)
	}
	return output, nil
}