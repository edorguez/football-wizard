package scraper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type headlessXResponse struct {
	URL  string `json:"url"`
	HTML string `json:"html"`
}

type Client struct {
	apiURL     string
	apiKey     string
	httpClient *http.Client
	logger     *slog.Logger
}

func NewClient(apiURL, apiKey string, logger *slog.Logger) *Client {
	return &Client{
		apiURL: apiURL,
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		logger: logger,
	}
}

func (c *Client) FetchHTML(url string) (string, error) {
	return c.fetchWithRetry(url, 2)
}

func (c *Client) fetchWithRetry(url string, attempts int) (string, error) {
	var lastErr error

	for i := range attempts {
		if i > 0 {
			c.logger.Info("retrying fetch", "url", url, "attempt", i+1)
			time.Sleep(time.Duration(i) * time.Second)
		}

		html, err := c.fetch(url)
		if err == nil {
			return html, nil
		}
		lastErr = err
	}

	return "", fmt.Errorf("fetching after %d attempts: %w", attempts, lastErr)
}

func (c *Client) fetch(url string) (string, error) {
	c.logger.Info("fetching page via HeadlessX", "url", url)

	body := map[string]string{"url": url}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/api/operators/website/scrape/html-js", c.apiURL),
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("requesting HeadlessX: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("HeadlessX returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var hxResp headlessXResponse
	if err := json.NewDecoder(resp.Body).Decode(&hxResp); err != nil {
		return "", fmt.Errorf("decoding HeadlessX response: %w", err)
	}

	c.logger.Info("page fetched via HeadlessX", "url", url, "size", len(hxResp.HTML))

	return hxResp.HTML, nil
}

func (c *Client) Close() error {
	return nil
}
