package domain

import "time"

type Direction string

const (
	In  Direction = "in"
	Out Direction = "out"
)

type Transaction struct {
	ID      string // serial primary key (ignored on insert)
	EmailID string // message ID from Mailpit or similar

	TxDate      time.Time // date extracted from email body
	TxBank      string    // e.g. "rbc"
	TxAccount   string    // e.g. "****1234"
	TxAmount    string    // "1234.56" (to preserve exact input for NUMERIC)
	TxCurrency  string    // e.g. "CAD"
	TxDirection Direction // "in" or "out"
	TxDesc      string    // raw transaction description (parsed from email)

	Category  string // to be AI-assigned later
	Merchant  string // inferred or parsed from description
	UserNotes string // manually entered by user later
}
