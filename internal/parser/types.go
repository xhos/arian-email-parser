package parser

import (
	"arian-parser/internal/domain"
	"strings"
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

// ToEmailMeta creates EmailMeta from decoded email content
func ToEmailMeta(id string, decodedContent string) (EmailMeta, error) {
	// Extract subject from email headers in the decoded content
	subject := extractSubject(decodedContent)
	// Extract date from email headers
	date := extractDate(decodedContent)

	return EmailMeta{
		ID:      id,
		Subject: subject,
		Text:    decodedContent,
		Date:    date,
	}, nil
}

// extractSubject finds the Subject header in decoded email content
func extractSubject(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Subject:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "Subject:"))
		}
	}
	return ""
}

// extractDate finds the Date header in decoded email content
func extractDate(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Date:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "Date:"))
		}
	}
	return ""
}
