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

	"arian-parser/internal/api"
	"arian-parser/internal/ingest"
	"arian-parser/internal/mailpit"

	"github.com/charmbracelet/log"
)

type Config struct {
	APIBaseURL  string
	APIKey      string
	WebhookPath string
	ListenAddr  string
}

func loadConfig() (Config, error) {
	cfg := Config{
		APIBaseURL:  os.Getenv("API_BASE_URL"),
		APIKey:      os.Getenv("API_KEY"),
		WebhookPath: os.Getenv("PARSER_WEBHOOK_PATH"),
		ListenAddr:  os.Getenv("PARSER_LISTEN_ADDR"),
	}

	if cfg.APIBaseURL == "" {
		return cfg, errors.New("API_BASE_URL must be set")
	}
	if cfg.APIKey == "" {
		return cfg, errors.New("API_KEY must be set")
	}

	// if webhook is not set, do a one-shot run
	if (cfg.WebhookPath == "") != (cfg.ListenAddr == "") {
		return cfg, errors.New("PARSER_WEBHOOK_PATH and PARSER_LISTEN_ADDR must both be set or both be empty")
	}

	return cfg, nil
}

func newProcessor(mp *mailpit.Client, apiClient *api.Client, accountMap map[string]int, lg *log.Logger) *ingest.Processor {
	return &ingest.Processor{
		MP:         mp,
		API:        apiClient,
		AccountMap: accountMap,
		Log:        lg.WithPrefix("proc"),
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

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
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

	// 1. initialize the API Client
	apiClient, err := api.NewClient()
	if err != nil {
		logger.Fatal("api client init", "err", err)
	}

	// 2. fetch accounts and build the lookup map
	logger.Info("fetching accounts from API to build lookup map")
	accounts, err := apiClient.GetAccounts()
	if err != nil {
		logger.Fatal("could not fetch accounts from backend", "err", err)
	}

	accountMap := make(map[string]int, len(accounts))
	logger.Info("building account map...")
	for _, acc := range accounts {
		if acc.Name == "" {
			logger.Warn("skipping account with empty name", "id", acc.ID)
			continue
		}
		key := fmt.Sprintf("%s-%s", acc.Bank, acc.Name)
		accountMap[key] = acc.ID
		logger.Info("added to map", "key", key, "id", acc.ID)
	}

	logger.Info("account map built", "count", len(accountMap))
	logger.Info("account map contents", "map", accountMap)

	// 3. initialize the Processor with new dependencies
	proc := newProcessor(mp, apiClient, accountMap, logger)

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
