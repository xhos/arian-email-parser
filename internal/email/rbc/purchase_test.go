package rbc

import (
	"path/filepath"
	"testing"
	"time"

	"arian-parser/internal/domain"
)

func TestPurchaseParser(t *testing.T) {
	assertTransaction(
		t,
		&purchase{},
		filepath.Join("testdata", "purchase.json"),
		"test-purchase-id",
		expectedTransactionDetails{
			Account:     "************0000",
			Amount:      "26.19",
			Date:        time.Date(2025, time.June, 1, 0, 0, 0, 0, time.UTC),
			Currency:    "CAD",
			Direction:   domain.Out,
			Description: "UBER EATS",
		},
	)
}

func TestPurchaseParserForwarded(t *testing.T) {
	assertTransaction(
		t,
		&purchase{},
		filepath.Join("testdata", "purchase_fwd.json"),
		"test-purchase-fwd-id",
		expectedTransactionDetails{
			Account:     "************0000",
			Amount:      "19.12",
			Date:        time.Date(2025, time.May, 25, 0, 0, 0, 0, time.UTC),
			Currency:    "CAD",
			Direction:   domain.Out,
			Description: "SOME STORE",
		},
	)
}
