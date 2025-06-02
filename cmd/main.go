package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"arian-parser/internal/db"
	"arian-parser/internal/ingest"
	"arian-parser/internal/mailpit"

	"github.com/charmbracelet/log"
)

type Config struct {
	PostgresURL string
	WebhookPath string
	ListenAddr  string
}

func loadConfig() (Config, error) {
	cfg := Config{
		PostgresURL: os.Getenv("POSTGRES_URL"),
		WebhookPath: os.Getenv("ARIAN_WEBHOOK_PATH"),
		ListenAddr:  os.Getenv("ARIAN_LISTEN_ADDR"),
	}

	if cfg.PostgresURL == "" {
		return cfg, errors.New("POSTGRES_URL must be set")
	}

	// if webhook is not set, do a one-shot run
	if (cfg.WebhookPath == "") != (cfg.ListenAddr == "") {
		return cfg, errors.New("ARIAN_WEBHOOK_PATH and ARIAN_LISTEN_ADDR must both be set or both be empty")
	}

	return cfg, nil
}

func newDB(dsn string, lg *log.Logger) (*db.DB, error) {
	lg.Info("connecting to postgres")
	return db.New(dsn)
}

func newProcessor(mp *mailpit.Client, dbConn *db.DB, lg *log.Logger) *ingest.Processor {
	return &ingest.Processor{
		MP:  mp,
		DB:  dbConn,
		Log: lg.WithPrefix("proc"),
	}
}

func newHTTPServer(addr, path string, proc *ingest.Processor, lg *log.Logger) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		_ = r.Body.Close() // we don't actually need the payload
		if err := proc.RunOnce(); err != nil {
			lg.Error("RunOnce failed", "err", err)
			http.Error(w, "processing failure", http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, "ok")
	})

	return &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
		ErrorLog:          log.NewWithOptions(os.Stderr, log.Options{Prefix: "http "}).StandardLog(),
	}
}

func main() {
	logger := log.NewWithOptions(os.Stdout, log.Options{Prefix: "arian"})

	cfg, err := loadConfig()
	if err != nil {
		logger.Fatal("config error", "err", err)
	}

	mp, err := mailpit.NewClient()
	if err != nil {
		logger.Fatal("mailpit init", "err", err)
	}

	dbConn, err := newDB(cfg.PostgresURL, logger)
	if err != nil {
		logger.Fatal("db init", "err", err)
	}
	defer dbConn.Close()

	proc := newProcessor(mp, dbConn, logger)

	// one-shot mode
	if cfg.WebhookPath == "" {
		logger.Info("running one-shot ingestion")
		if err := proc.RunOnce(); err != nil {
			logger.Fatal("RunOnce failed", "err", err)
		}

		logger.Info("done")
		return
	}

	// webhook mode
	srv := newHTTPServer(cfg.ListenAddr, cfg.WebhookPath, proc, logger)
	go func() {
		logger.Info("http listen", "addr", cfg.ListenAddr, "path", cfg.WebhookPath)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("http server error", "err", err)
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	logger.Info("shutting down. bye!")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
