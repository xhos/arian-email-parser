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
create table if not exists transactions (
  id serial primary key,

  email_id          text     not null,
  email_received_at timestamptz not null,

  tx_date      timestamptz   not null,
  tx_bank      text          not null,
  tx_account   text          not null,
  tx_amount    numeric(18,2) not null,
  tx_currency  text          not null,
  tx_direction text          not null,
  tx_desc      text,

  category   text,
  merchant   text,
  user_notes text
);

create unique index if not exists idx_transactions_email_id
  on transactions(email_id);

create index if not exists idx_transactions_tx_date
  on transactions(tx_date);
`
	_, err := w.Exec(ddl)
	return err
}

// Insert inserts a transaction into the DB, skipping duplicates
func (w *DB) Insert(tx *domain.Transaction) error {
	const q = `
insert into transactions (
  id,
  email_id,
  email_received_at,
  tx_date,
  tx_bank,
  tx_account,
  tx_amount,
  tx_currency,
  tx_direction,
  tx_desc,
  category,
  merchant,
  user_notes
)
values (
  default,
  :email_id,
  :email_received_at,
  :tx_date,
  :tx_bank,
  :tx_account,
  :tx_amount,
  :tx_currency,
  :tx_direction,
  :tx_desc,
  :category,
  :merchant,
  :user_notes
);`

	data := map[string]any{
		"email_id":          tx.EmailID,
		"email_received_at": tx.EmailReceivedAt,
		"tx_date":           tx.TxDate,
		"tx_bank":           tx.TxBank,
		"tx_account":        tx.TxAccount,
		"tx_amount":         tx.TxAmount,
		"tx_currency":       tx.TxCurrency,
		"tx_direction":      string(tx.TxDirection),
		"tx_desc":           tx.TxDesc,
		"category":          tx.Category,
		"merchant":          tx.Merchant,
		"user_notes":        tx.UserNotes,
	}

	_, err := w.NamedExec(q, data)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			w.log.Info("duplicate email_id, skipping insert", "email_id", tx.EmailID)
			return nil
		}
		w.log.Warn("insert failed", "email_id", tx.EmailID, "err", err)
		return err
	}
	return nil
}

// Close shuts down the DB connection (logs a message, returns any error)
func (w *DB) Close() error {
	w.log.Info("closing DB")
	return w.DB.Close()
}
