package rbc

import (
	"regexp"
	"strings"

	"arian-parser/internal/domain"
	"arian-parser/internal/parser"
)

func init() { parser.Register(&credit{}) }

type credit struct{}

func (p *credit) Match(m parser.EmailMeta) bool {
	return strings.Contains(m.Subject, "You received a credit.") &&
		strings.Contains(m.Text, "RBC Royal Bank")
}

func (p *credit) Parse(m parser.EmailMeta) (*domain.Transaction, error) {
	patterns := map[string]*regexp.Regexp{
		"account": regexp.MustCompile(`Account:(?:\s*\n\s*>\s*\n\s*>\s*|\s*)(\*+\d+)`),
		"amount":  regexp.MustCompile(`Purchase Amount:(?:\s*\n\s*>\s*\n\s*>\s*|\s*)\$([\d,]+\.\d{2})`),
		"txdate":  regexp.MustCompile(`Transaction Date:(?:\s*\n\s*>\s*\n\s*>\s*|\s*)([A-Za-z]+ \d{1,2}, \d{4})`),
		"desc":    regexp.MustCompile(`Transaction Description:(?:\s*\n\s*>\s*\n\s*>\s*|\s*)(.+)`),
	}
	fields, err := parser.ExtractFields(m.Text, patterns)
	if err != nil {
		return nil, err
	}

	return parser.BuildTransaction(
		m,
		fields,
		"rbc",
		"CAD",
		domain.In,
		strings.TrimSpace(fields["desc"]),
	)
}
