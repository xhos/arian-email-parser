package rbc

import (
	"path/filepath"
	"testing"
	"time"

	"arian-parser/internal/domain"
)

func TestWithdrawalParserForwarded(t *testing.T) {
	assertTransaction(
		t,
		&withdrawal{},
		filepath.Join("testdata", "withdrawal_fwd.txt"),
		"test-withdrawal-fwd-id",
		expectedTransactionDetails{
			Account:     "TestChecking",
			Amount:      "75.50",
			Date:        time.Date(2025, time.August, 2, 0, 0, 0, 0, time.UTC),
			Currency:    "CAD",
			Direction:   domain.Out,
			Description: "RBC Withdrawal",
		},
	)
}
