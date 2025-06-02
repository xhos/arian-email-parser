package db

import (
	"fmt"
	"strings"

	"arian-parser/internal/domain"

	"github.com/charmbracelet/log"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type DB struct {
	*sqlx.DB
	log *log.Logger
}

// New opens a connection, pings it, ensures schema, and returns *DB
func New(dsn string) (*DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("empty DSN")
	}

	dbConn, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlx.Open: %w", err)
	}

	if err := dbConn.Ping(); err != nil {
		return nil, fmt.Errorf("db ping: %w", err)
	}

	w := &DB{
		DB:  dbConn,
		log: log.WithPrefix("db"),
	}

	if err := w.ensureSchema(); err != nil {
		_ = w.Close()
		return nil, fmt.Errorf("ensureSchema: %w", err)
	}
	return w, nil
}

// ensureSchema creates the transactions table (if needed), adds a unique index on email_id and indexes txn_date for faster queries.
func (w *DB) ensureSchema() error {
	const ddl = `
CREATE TABLE IF NOT EXISTS transactions (
    id           SERIAL PRIMARY KEY,
    bank         TEXT NOT NULL,
    email_id     TEXT NOT NULL,
    received_at  TIMESTAMPTZ NOT NULL,
    txn_date     TIMESTAMPTZ NOT NULL,
    account      TEXT NOT NULL,
    amount       NUMERIC(18,2) NOT NULL,
    currency     TEXT NOT NULL,
    direction    TEXT NOT NULL,
    category     TEXT,
    merchant     TEXT NOT NULL,
    raw          TEXT NOT NULL,
    meta         JSONB DEFAULT '{}'::jsonb
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_email_id
    ON transactions(email_id);

CREATE INDEX IF NOT EXISTS idx_transactions_txn_date
    ON transactions(txn_date);
`
	_, err := w.Exec(ddl)
	return err
}

// Insert inserts a transaction into the DB. If email_id already exists, it returns ErrEmailExists
func (w *DB) Insert(tx *domain.Transaction) error {
	const q = `
INSERT INTO transactions
  (bank, email_id, received_at, txn_date, account, amount, currency,
   direction, category, merchant, raw, meta)
VALUES
  (:bank, :email_id, :received_at, :txn_date, :account, :amount, :currency,
   :direction, :category, :merchant, :raw, :meta);
`

	data := map[string]any{
		"bank":        tx.Bank,
		"email_id":    tx.EmailID,
		"received_at": tx.ReceivedAt,
		"txn_date":    tx.TxnDate,
		"account":     tx.Account,
		"amount":      tx.Amount,
		"currency":    tx.Currency,
		"direction":   string(tx.Direction),
		"category":    tx.Category,
		"merchant":    tx.Merchant,
		"raw":         tx.Raw,
		"meta":        tx.Meta,
	}

	_, err := w.NamedExec(q, data)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrEmailExists
		}
		w.log.Warn("insert failed", "err", err)
		return err
	}
	return nil
}

// ErrEmailExists indicates a duplicate email_id
var ErrEmailExists = fmt.Errorf("email_id already exists")

// isUniqueViolation checks if the error is a UNIQUE constraint violation
func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "duplicate key value violates unique constraint")
}

// Close shuts down the DB connection (logs a message, returns any error)
func (w *DB) Close() error {
	w.log.Info("closing DB")
	return w.DB.Close()
}
