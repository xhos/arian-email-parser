package mailpit

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/charmbracelet/log"
)

// Client is a simple HTTP client for interacting with the Mailpit API
type Client struct {
	httpClient *http.Client
	baseURL    *url.URL
	authHeader string
	log        *log.Logger
}

// NewClient initializes a new Mailpit client using environment variables
func NewClient() (*Client, error) {
	raw := os.Getenv("MAILPIT_URL")
	if raw == "" {
		return nil, fmt.Errorf("MAILPIT_URL not set")
	}

	user, pass := os.Getenv("MAILPIT_USERNAME"), os.Getenv("MAILPIT_PASSWORD")
	if user == "" || pass == "" {
		return nil, fmt.Errorf("MAILPIT_USERNAME and MAILPIT_PASSWORD must be set")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid MAILPIT_URL: %w", err)
	}

	u.Path = path.Clean(u.Path) + "/"

	return &Client{
		httpClient: http.DefaultClient,
		baseURL:    u,
		authHeader: "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+pass)),
		log:        log.NewWithOptions(os.Stderr, log.Options{Prefix: "mailpit"}),
	}, nil
}

// doRequest performs an HTTP request to the Mailpit API
func (c *Client) doRequest(method, rel string) ([]byte, error) {
	u := *c.baseURL
	u.Path = path.Join(c.baseURL.Path, rel)

	req, err := http.NewRequestWithContext(context.Background(), method, u.String(), nil)
	if err != nil {
		c.log.Error("create request", "err", err)
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	if c.authHeader != "" {
		req.Header.Set("Authorization", c.authHeader)
	}

	c.log.Debug("http", "method", method, "url", u.String())

	res, err := c.httpClient.Do(req)
	if err != nil {
		c.log.Error("send request", "err", err)
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(res.Body)

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		err = fmt.Errorf("%s: %s", res.Status, body)
		c.log.Error("bad status", "status", res.Status, "body", string(body))
		return nil, err
	}

	return io.ReadAll(res.Body)
}
