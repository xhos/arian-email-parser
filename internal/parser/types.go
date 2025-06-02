package parser

import (
	"arian-parser/internal/domain"
	"encoding/json"
)

// EmailMeta holds the minimal fields needed to pick and parse a message
type EmailMeta struct {
	ID      string
	Subject string
	Text    string
	Date    string // RFC3339 from Mailpit
}

// Parser defines a bank‐specific parser
type Parser interface {
	Match(meta EmailMeta) bool
	Parse(meta EmailMeta) (*domain.Transaction, error)
}

// toEmailMeta unmarshals the raw JSON from Mailpit into EmailMeta
func ToEmailMeta(id string, raw []byte) (EmailMeta, error) {
	var m struct {
		Subject string `json:"Subject"`
		Text    string `json:"Text"`
		Date    string `json:"Date"`
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		return EmailMeta{}, err
	}
	return EmailMeta{
		ID:      id,
		Subject: m.Subject,
		Text:    m.Text,
		Date:    m.Date,
	}, nil
}
