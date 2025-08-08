package rbc

import (
	"path/filepath"
	"testing"
	"time"

	"arian-parser/internal/domain"
)

func TestCreditParserForwarded(t *testing.T) {
	assertTransaction(
		t,
		&credit{},
		filepath.Join("testdata", "credit_fwd.txt"),
		"test-credit-fwd-id",
		expectedTransactionDetails{
			Account:     "************1234",
			Amount:      "123.45",
			Date:        time.Date(2025, time.June, 10, 0, 0, 0, 0, time.UTC),
			Currency:    "CAD",
			Direction:   domain.In,
			Description: "TEST MERCHANT CORP",
		},
	)
}
