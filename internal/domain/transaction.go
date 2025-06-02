package domain

import "time"

type Direction string

const (
	In  Direction = "in"
	Out Direction = "out"
)

type Transaction struct {
	ID         string    // db primary key (set on insert)
	Bank       string    // "rbc"
	EmailID    string    // Mailpit message ID
	ReceivedAt time.Time // when email arrived
	TxnDate    time.Time // date reported in email
	Account    string
	Amount     string
	Currency   string
	Direction  Direction // in / out
	Category   string    // later AI-filled
	Merchant   string
	Raw        string            // entire email body or hash
	Meta       map[string]string // sparse extras (sender, etc.)
}
