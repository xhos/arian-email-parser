package ingest

import (
	"arian-parser/internal/api"
	_ "arian-parser/internal/email/all"
	"arian-parser/internal/mailpit"
	"arian-parser/internal/parser"
	"fmt"
	"github.com/charmbracelet/log"
	"strings"
)

// Processor glues Mailpit -> parser -> API
type Processor struct {
	MP         *mailpit.Client
	API        *api.Client
	AccountMap map[string]int
	Log        *log.Logger
}

// RunOnce pulls unread mails, parses, and stores them
func (p *Processor) RunOnce() error {
	ids, err := p.MP.UnreadIDs()
	if err != nil {
		return err
	}

	for _, id := range ids {
		raw, err := p.MP.Message(id)
		if err != nil {
			p.Log.Warn("fetch failed", "id", id, "err", err)
			continue
		}

		meta, err := parser.ToEmailMeta(id, raw)
		if err != nil {
			p.Log.Warn("meta parse", "id", id, "err", err)
			continue
		}

		prsr := parser.Find(meta)
		if prsr == nil {
			p.Log.Info("no parser matched", "id", id, "subject", meta.Subject)
			continue
		}

		txn, err := prsr.Parse(meta)
		if err != nil {
			p.Log.Info("email text", "id", id, "text", meta.Text)
			p.Log.Error("parse failed", "id", id, "err", err)
			continue
		}
		if txn == nil {
			continue
		}

		cleanAccount := strings.TrimLeft(txn.TxAccount, "*")
		accountKey := fmt.Sprintf("%s-%s", txn.TxBank, cleanAccount)
		accountID, ok := p.AccountMap[accountKey]
		if !ok {
			p.Log.Warn("unrecognized account; skipping",
				"id", id,
				"bank", txn.TxBank,
				"account", cleanAccount)
			continue
		}

		txn.AccountID = accountID

		p.Log.Info("sending to api", "id", id, "account_id", accountID)
		if err := p.API.CreateTransaction(txn); err != nil {
			p.Log.Error("api create transaction failed", "id", id, "err", err)
		}
	}

	return nil
}
