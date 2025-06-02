package rbc

import (
	"path/filepath"
	"testing"
	"time"

	"arian-parser/internal/domain"
)

func TestWithdrawalParser(t *testing.T) {
	assertTransaction(
		t,
		&withdrawal{},
		filepath.Join("testdata", "withdrawal.json"),
		"test-withdrawal-id",
		expectedTransactionDetails{
			Account:     "Savings",
			Amount:      "2200.00",
			Date:        time.Date(2025, time.June, 1, 0, 0, 0, 0, time.UTC),
			Currency:    "CAD",
			Direction:   domain.Out,
			Description: "RBC Withdrawal",
		},
	)
}

func TestWithdrawalParserForwarded(t *testing.T) {
	assertTransaction(
		t,
		&withdrawal{},
		filepath.Join("testdata", "withdrawal_fwd.json"),
		"test-withdrawal-fwd-id",
		expectedTransactionDetails{
			Account:     "Daily",
			Amount:      "50.00",
			Date:        time.Date(2025, time.May, 28, 0, 0, 0, 0, time.UTC),
			Currency:    "CAD",
			Direction:   domain.Out,
			Description: "RBC Withdrawal",
		},
	)
}
