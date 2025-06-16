package rbc

import (
	"regexp"
	"strings"

	"arian-parser/internal/domain"
	"arian-parser/internal/parser"
)

func init() { parser.Register(&payment{}) }

type payment struct{}

func (p *payment) Match(m parser.EmailMeta) bool {
	return strings.Contains(m.Subject, "Payment Made") &&
		strings.Contains(m.Text, "RBC Royal Bank")
}

func (p *payment) Parse(m parser.EmailMeta) (*domain.Transaction, error) {
	patterns := map[string]*regexp.Regexp{
		"account": regexp.MustCompile(`(?s)[>\s]*Account:.*?(\*+\d+)`),
		"amount":  regexp.MustCompile(`(?s)[>\s]*Payment Amount:.*?\$\s*([\d,]+\.\d{2})`),
		"txdate":  regexp.MustCompile(`(?s)[>\s]*Transaction Date:.*?([A-Za-z]+\s+\d{1,2},\s+\d{4})`),
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
		"RBC Payment",
	)
}
