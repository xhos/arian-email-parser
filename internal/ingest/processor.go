package ingest

import (
	"arian-parser/internal/domain"
	_ "arian-parser/internal/email/all"
	"arian-parser/internal/mailpit"
	"arian-parser/internal/parser"

	"github.com/charmbracelet/log"
)

type dbWriter interface {
	Insert(tx *domain.Transaction) error
}

// Processor glues Mailpit -> parser -> DB
type Processor struct {
	MP  *mailpit.Client
	DB  dbWriter
	Log *log.Logger
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

		p.Log.Info("insert", "id", id)
		if err := p.DB.Insert(txn); err != nil {
			p.Log.Error("db insert", "id", id, "err", err)
		}
	}

	return nil
}
