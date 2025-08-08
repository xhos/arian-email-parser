package rbc

import (
	"path/filepath"
	"testing"
	"time"

	"arian-parser/internal/domain"
)

func TestDepositParserForwarded(t *testing.T) {
	assertTransaction(
		t,
		&deposit{},
		filepath.Join("testdata", "deposit_fwd.txt"),
		"test-deposit-fwd-id",
		expectedTransactionDetails{
			Account:     "TestSavings",
			Amount:      "50.25",
			Date:        time.Date(2025, time.August, 5, 0, 0, 0, 0, time.UTC),
			Currency:    "CAD",
			Direction:   domain.In,
			Description: "RBC Deposit",
		},
	)
}
