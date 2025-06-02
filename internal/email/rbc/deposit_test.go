package rbc

import (
	"path/filepath"
	"testing"
	"time"

	"arian-parser/internal/domain"
)

func TestDepositParser(t *testing.T) {
	assertTransaction(
		t,
		&deposit{},
		filepath.Join("testdata", "deposit.json"),
		"test-deposit-id",
		expectedTransactionDetails{
			Account:   "Savings",
			Amount:    "42000.16",
			Date:      time.Date(2025, time.May, 31, 0, 0, 0, 0, time.UTC),
			Currency:  "CAD",
			Direction: domain.In,
			Merchant:  "RBC Deposit",
		},
	)
}

func TestDepositParserForwarded(t *testing.T) {
	assertTransaction(
		t,
		&deposit{},
		filepath.Join("testdata", "deposit_fwd.json"),
		"test-deposit-fwd-id",
		expectedTransactionDetails{
			Account:   "Daily",
			Amount:    "60.00",
			Date:      time.Date(2025, time.May, 27, 0, 0, 0, 0, time.UTC),
			Currency:  "CAD",
			Direction: domain.In,
			Merchant:  "RBC Deposit",
		},
	)
}
