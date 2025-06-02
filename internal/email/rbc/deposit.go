package rbc

import (
	"regexp"
	"strings"

	"arian-parser/internal/domain"
	"arian-parser/internal/parser"
)

func init() { parser.Register(&deposit{}) }

type deposit struct{}

func (d *deposit) Match(m parser.EmailMeta) bool {
	return strings.Contains(m.Subject, "Deposit Notice") &&
		strings.Contains(m.Text, "RBC Royal Bank")
}

func (d *deposit) Parse(m parser.EmailMeta) (*domain.Transaction, error) {
	patterns := map[string]*regexp.Regexp{
		"account": regexp.MustCompile(`Account:(?:\s*\n\s*>\s*\n\s*>\s*|\s+)([A-Za-z0-9 ]+)`),
		"amount":  regexp.MustCompile(`Deposit Amount:(?:\s*\n\s*>\s*\n\s*>\s*|\s+)\$([\d,]+\.\d{2})`),
		"txdate":  regexp.MustCompile(`Transaction Date:(?:\s*\n\s*>\s*\n\s*>\s*|\s+)([A-Za-z]+ \d{1,2}, \d{4})`),
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
		"RBC Deposit",
	)
}
