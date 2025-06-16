package rbc

import (
	"path/filepath"
	"testing"
	"time"

	"arian-parser/internal/domain"
)

func TestCreditParser(t *testing.T) {
	assertTransaction(
		t,
		&credit{},
		filepath.Join("testdata", "credit.json"),
		"test-credit-id",
		expectedTransactionDetails{
			Account:     "************0000",
			Amount:      "840.72",
			Date:        time.Date(2025, time.June, 10, 17, 43, 39, 0, time.FixedZone("UTC-6", -6*60*60)),
			Currency:    "CAD",
			Direction:   domain.In,
			Description: "CREDIT GIVER",
		},
	)
}

func TestCreditParserForwarded(t *testing.T) {
	assertTransaction(
		t,
		&credit{},
		filepath.Join("testdata", "credit_fwd.json"),
		"test-credit-fwd-id",
		expectedTransactionDetails{
			Account:     "************0000",
			Amount:      "840.72",
			Date:        time.Date(2025, time.June, 10, 0, 0, 0, 0, time.UTC),
			Currency:    "CAD",
			Direction:   domain.In,
			Description: "CREDIT GIVER",
		},
	)
}
