package rbc

import (
	"os"
	"testing"
	"time"

	"arian-parser/internal/domain"
	"arian-parser/internal/parser"
)

type expectedTransactionDetails struct {
	Account   string
	Amount    string
	Date      time.Time
	Currency  string
	Direction domain.Direction
	Merchant  string
}

func assertTransaction(
	t *testing.T,
	p parser.Parser,
	fixturePath string,
	emailID string,
	expected expectedTransactionDetails,
) {
	t.Helper()

	rawBytes, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", fixturePath, err)
	}

	meta, err := parser.ToEmailMeta(emailID, rawBytes)
	if err != nil {
		t.Fatalf("toEmailMeta failed for %s: %v", fixturePath, err)
	}

	if !p.Match(meta) {
		t.Fatalf("Match(meta) was false for %s; subject=%q", fixturePath, meta.Subject)
	}

	txn, err := p.Parse(meta)
	if err != nil {
		t.Fatalf("Parse(meta) returned error for %s: %v", fixturePath, err)
	}
	if txn == nil {
		t.Fatalf("Parse(meta) returned nil Transaction for %s but expected a real one", fixturePath)
	}

	if txn.Account != expected.Account {
		t.Errorf("Account = %q; want %q (fixture: %s)", txn.Account, expected.Account, fixturePath)
	}
	if txn.Amount != expected.Amount {
		t.Errorf("Amount = %q; want %q (fixture: %s)", txn.Amount, expected.Amount, fixturePath)
	}
	if !txn.TxnDate.Equal(expected.Date) {
		t.Errorf("TxnDate = %v; want %v (fixture: %s)", txn.TxnDate, expected.Date, fixturePath)
	}
	if txn.Currency != expected.Currency {
		t.Errorf("Currency = %q; want %q (fixture: %s)", txn.Currency, expected.Currency, fixturePath)
	}
	if txn.Direction != expected.Direction {
		t.Errorf("Direction = %v; want %v (fixture: %s)", txn.Direction, expected.Direction, fixturePath)
	}
	if txn.Merchant != expected.Merchant {
		t.Errorf("Merchant = %q; want %q (fixture: %s)", txn.Merchant, expected.Merchant, fixturePath)
	}
}
