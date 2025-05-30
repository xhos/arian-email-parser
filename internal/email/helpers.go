package email

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/charmbracelet/log"
)

type Client struct {
	httpClient *http.Client
	baseURL    *url.URL
	authHeader string
	log        *log.Logger
}

func NewClient() (*Client, error) {
	raw := os.Getenv("MAILPIT_URL")
	if raw == "" {
		return nil, errors.New("MAILPIT_URL not set")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid MAILPIT_URL: %w", err)
	}

	u.Path = path.Clean(u.Path) + "/"

	user, pass := os.Getenv("MAILPIT_USERNAME"), os.Getenv("MAILPIT_PASSWORD")
	var auth string
	if user != "" && pass != "" {
		auth = "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+pass))
	}

	logger := log.NewWithOptions(os.Stderr, log.Options{Prefix: "mailpit"})

	return &Client{http.DefaultClient, u, auth, logger}, nil
}

func (c *Client) doRequest(ctx context.Context, method, relPath string) ([]byte, error) {
	u := *c.baseURL
	u.Path = path.Join(c.baseURL.Path, relPath)

	req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		c.log.Error("create request failed", "error", err)
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	if c.authHeader != "" {
		req.Header.Set("Authorization", c.authHeader)
	}

	c.log.Debug("request", "method", method, "url", u.String())

	res, err := c.httpClient.Do(req)
	if err != nil {
		c.log.Error("request failed", "error", err)
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		err = fmt.Errorf("%s: %s", res.Status, body)
		c.log.Error("bad response", "status", res.Status, "body", string(body))
		return nil, err
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		c.log.Error("read body failed", "error", err)
		return nil, err
	}

	c.log.Debug("response received", "length", len(data))

	return data, nil
}
