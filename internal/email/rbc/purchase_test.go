package rbc

import (
	"path/filepath"
	"testing"
	"time"

	"arian-parser/internal/domain"
)

func TestPurchaseParserForwarded(t *testing.T) {
	assertTransaction(
		t,
		&purchase{},
		filepath.Join("testdata", "purchase_fwd.txt"),
		"test-purchase-fwd-id",
		expectedTransactionDetails{
			Account:     "************5678",
			Amount:      "67.89",
			Date:        time.Date(2025, time.August, 5, 0, 0, 0, 0, time.UTC),
			Currency:    "CAD",
			Direction:   domain.Out,
			Description: "TEST GROCERY STORE",
		},
	)
}
