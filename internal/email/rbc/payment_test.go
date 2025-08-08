package rbc

import (
	"path/filepath"
	"testing"
	"time"

	"arian-parser/internal/domain"
)

func TestPaymentParserForwarded(t *testing.T) {
	assertTransaction(
		t,
		&payment{},
		filepath.Join("testdata", "payment_fwd.txt"),
		"test-payment-fwd-id",
		expectedTransactionDetails{
			Account:     "************4567",
			Amount:      "250.00",
			Date:        time.Date(2025, time.August, 1, 0, 0, 0, 0, time.UTC),
			Currency:    "CAD",
			Direction:   domain.In,
			Description: "RBC Payment",
		},
	)
}
