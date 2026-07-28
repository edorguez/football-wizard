package scraper

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/stealth"
)

type Client struct {
	browser *rod.Browser
	logger  *slog.Logger
}

func NewClient(logger *slog.Logger) (*Client, error) {
	path, has := launcher.LookPath()
	if !has {
		return nil, fmt.Errorf("chrome/chromium not found on system path")
	}

	url := launcher.New().
		Bin(path).
		Headless(true).
		NoSandbox(true).
		MustLaunch()

	browser := rod.New().
		ControlURL(url).
		Timeout(30 * time.Second).
		MustConnect()

	return &Client{
		browser: browser,
		logger:  logger,
	}, nil
}

func (c *Client) FetchHTML(url string) (string, error) {
	c.logger.Info("fetching page", "url", url)

	page, err := stealth.Page(c.browser)
	if err != nil {
		return "", fmt.Errorf("creating stealth page: %w", err)
	}
	defer func() {
		if err := page.Close(); err != nil {
			c.logger.Error("closing page", "error", err)
		}
	}()

	if err := page.Navigate(url); err != nil {
		return "", fmt.Errorf("navigating to %s: %w", url, err)
	}

	if err := page.WaitLoad(); err != nil {
		return "", fmt.Errorf("waiting for page load: %w", err)
	}

	html, err := page.HTML()
	if err != nil {
		return "", fmt.Errorf("getting page HTML: %w", err)
	}

	c.logger.Info("page fetched", "url", url, "size", len(html))

	return html, nil
}

func (c *Client) Close() error {
	return c.browser.Close()
}
