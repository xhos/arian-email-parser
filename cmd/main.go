package main

import (
	"os"

	"arian-parser/internal/db"
	"arian-parser/internal/ingest"
	"arian-parser/internal/mailpit"

	"github.com/charmbracelet/log"
)

func main() {
	logger := log.NewWithOptions(os.Stdout, log.Options{Prefix: "arian"})

	mp, err := mailpit.NewClient()
	if err != nil {
		logger.Fatal("mailpit init", "err", err)
	}

	dbConn, err := db.New(os.Getenv("POSTGRES_URL"))
	if err != nil {
		logger.Fatal("db init", "err", err)
	}

	proc := &ingest.Processor{
		MP:  mp,
		DB:  dbConn,
		Log: logger,
	}

	if err := proc.RunOnce(); err != nil {
		logger.Error("processor", "err", err)
	}
}
