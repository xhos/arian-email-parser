package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"arian-parser/internal/domain"

	"github.com/charmbracelet/log"
)

type Account struct {
	ID            int     `json:"id"`
	Alias         string  `json:"alias"`
	Bank          string  `json:"bank"`
	Name          string  `json:"name"`
	Type          string  `json:"type"`
	AnchorBalance float64 `json:"anchor_balance"`
	AnchorDate    string  `json:"anchor_date"`
	CreatedAt     string  `json:"created_at"`
}

type CreateTransactionRequest struct {
	AccountID   int     `json:"account_id"`
	EmailID     string  `json:"email_id"`
	TxDate      string  `json:"tx_date"`
	TxAmount    float64 `json:"tx_amount"`
	TxDirection string  `json:"tx_direction"`
	TxDesc      string  `json:"tx_desc"`
	TxCurrency  string  `json:"tx_currency"`
	Merchant    string  `json:"merchant,omitempty"`
	Category    string  `json:"category,omitempty"`
	UserNotes   string  `json:"user_notes,omitempty"`
}

type Client struct {
	httpClient *http.Client
	baseURL    *url.URL
	apiKey     string
	log        *log.Logger
}

func NewClient() (*Client, error) {
	rawURL := os.Getenv("API_BASE_URL")
	if rawURL == "" {
		return nil, fmt.Errorf("API_BASE_URL not set")
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("API_KEY not set")
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid API_BASE_URL: %w", err)
	}

	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    u,
		apiKey:     apiKey,
		log:        log.NewWithOptions(os.Stderr, log.Options{Prefix: "api-client"}),
	}, nil
}

func (c *Client) GetAccounts() ([]Account, error) {
	body, statusCode, err := c.doRequest(http.MethodGet, "/api/accounts", nil)
	if err != nil {
		return nil, fmt.Errorf("request to get accounts failed: %w", err)
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get accounts, received status %d: %s", statusCode, string(body))
	}

	var accounts []Account
	if err := json.Unmarshal(body, &accounts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal accounts response: %w", err)
	}

	c.log.Info("successfully fetched accounts", "count", len(accounts))
	return accounts, nil
}

func (c *Client) CreateTransaction(tx *domain.Transaction) error {
	apiRequest := CreateTransactionRequest{
		AccountID:   tx.AccountID,
		EmailID:     tx.EmailID,
		TxDate:      tx.TxDate.Format(time.RFC3339),
		TxAmount:    tx.TxAmount,
		TxDirection: string(tx.TxDirection),
		TxDesc:      tx.TxDesc,
		TxCurrency:  tx.TxCurrency,
		Merchant:    tx.Merchant,
		Category:    tx.Category,
		UserNotes:   tx.UserNotes,
	}

	payload, err := json.Marshal(apiRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction payload: %w", err)
	}

	body, statusCode, err := c.doRequest(http.MethodPost, "/api/transactions", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("request to create transaction failed: %w", err)
	}

	// handle API-specific errors based on status code
	switch statusCode {
	case http.StatusCreated:
		c.log.Info("transaction created successfully", "email_id", tx.EmailID)
		return nil
	case http.StatusConflict:
		c.log.Info("skipping duplicate transaction", "email_id", tx.EmailID)
		return nil // not a fatal error, just a duplicate
	default:
		return fmt.Errorf("failed to create transaction, status %d: %s", statusCode, string(body))
	}
}

// doRequest is a helper that performs an http request and handles authentication
func (c *Client) doRequest(method, relPath string, payload io.Reader) ([]byte, int, error) {
	u := *c.baseURL
	u.Path = path.Join(c.baseURL.Path, relPath)

	req, err := http.NewRequestWithContext(context.Background(), method, u.String(), payload)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	c.log.Debug("sending request", "method", method, "url", u.String())
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, res.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, res.StatusCode, nil
}
