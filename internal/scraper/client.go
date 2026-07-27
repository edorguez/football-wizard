package scraper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ProgressMsg struct {
	Text string
}

type Client struct {
	apiURL   string
	apiKey   string
	http     *http.Client
	Progress chan ProgressMsg
}

func NewClient(apiURL, apiKey string) *Client {
	return &Client{
		apiURL:   apiURL,
		apiKey:   apiKey,
		http:     &http.Client{Timeout: 60 * time.Second},
		Progress: make(chan ProgressMsg, 20),
	}
}

func (c *Client) FetchHTML(url string) (string, error) {
	c.Progress <- ProgressMsg{Text: fmt.Sprintf("sending URL to HeadlessX browser: %s", url)}

	body, _ := json.Marshal(map[string]string{"url": url})
	req, _ := http.NewRequest("POST", c.apiURL+"/api/operators/website/scrape/html-js", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		c.Progress <- ProgressMsg{Text: fmt.Sprintf("ERROR: HeadlessX request failed — %s", err)}
		return "", fmt.Errorf("headlessx request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		c.Progress <- ProgressMsg{Text: fmt.Sprintf("ERROR: HeadlessX returned HTTP %d", resp.StatusCode)}
		return "", fmt.Errorf("headlessx returned HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		HTML string `json:"html"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		c.Progress <- ProgressMsg{Text: fmt.Sprintf("ERROR: parsing HeadlessX response — %s", err)}
		return "", fmt.Errorf("parsing headlessx response: %w", err)
	}

	c.Progress <- ProgressMsg{Text: fmt.Sprintf("HTML received — %d bytes", len(result.HTML))}
	return result.HTML, nil
}
