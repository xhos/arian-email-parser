package parser

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"arian-parser/internal/domain"
)

// ExtractFields applies each regex to the email body and returns the single capture group for each key
func ExtractFields(emailBody string, patterns map[string]*regexp.Regexp) (map[string]string, error) {
	out := make(map[string]string, len(patterns))
	for key, re := range patterns {
		m := re.FindStringSubmatch(emailBody)
		if len(m) < 2 {
			return nil, fmt.Errorf("field %q not found in text", key)
		}
		out[key] = m[1]
	}

	return out, nil
}

// ParseEmailDate normalizes a string into time.Time
func ParseEmailDate(raw string) (time.Time, error) {
	return time.Parse("January 2, 2006", raw)
}

// BuildTransaction assembles a domain.Transaction
func BuildTransaction(
	m EmailMeta,
	fields map[string]string,
	bank, currency string,
	dir domain.Direction,
	merchant string,
) (*domain.Transaction, error) {
	// 1. parse ReceivedAt (RFC3339 from Mailpit)
	recv, err := time.Parse(time.RFC3339, m.Date)
	if err != nil {
		return nil, err
	}

	// 2. parse the "tx date" from the email body
	txDate, err := ParseEmailDate(fields["txdate"])
	if err != nil {
		return nil, err
	}

	// 3. strip commas from the captured amount
	rawAmt := strings.ReplaceAll(fields["amount"], ",", "")

	return &domain.Transaction{
		Bank:       bank,
		EmailID:    m.ID,
		ReceivedAt: recv,
		TxnDate:    txDate,
		Account:    fields["account"],
		Amount:     rawAmt,
		Currency:   currency,
		Direction:  dir,
		Merchant:   merchant,
		Raw:        m.Text,
		Meta:       nil,
	}, nil
}
