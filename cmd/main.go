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
	"arian-parser/internal/smtp"

	"github.com/charmbracelet/log"
)

type Config struct {
	AriandURL  string
	APIKey     string
	SMTPAddr   string
	SMTPDomain string
	TLSCert    string
	TLSKey     string
}

func loadConfig() (Config, error) {
	cfg := Config{
		AriandURL:  os.Getenv("ARIAND_URL"),
		APIKey:     os.Getenv("API_KEY"),
		SMTPAddr:   getEnvDefault("SMTP_ADDR", ":2525"),
		SMTPDomain: getEnvDefault("SMTP_DOMAIN", "localhost"),
		TLSCert:    os.Getenv("TLS_CERT"),
		TLSKey:     os.Getenv("TLS_KEY"),
	}

	if cfg.AriandURL == "" {
		return cfg, errors.New("ARIAND_URL must be set")
	}
	if cfg.APIKey == "" {
		return cfg, errors.New("API_KEY must be set")
	}

	return cfg, nil
}

func getEnvDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func newHTTPServer(addr string, _ *log.Logger) *http.Server {
	mux := http.NewServeMux()
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

	// 1. initialize the API Client
	apiClient, err := api.NewClient(cfg.AriandURL, "", cfg.APIKey)
	if err != nil {
		logger.Fatal("api client init", "err", err)
	}
	defer func() {
		if err := apiClient.Close(); err != nil {
			logger.Error("failed to close gRPC connection", "err", err)
		}
	}()

	// 2. initialize the email handler and smtp server
	handler := smtp.NewEmailHandler(apiClient, logger)
	smtpServer := smtp.NewServer(cfg.SMTPAddr, cfg.SMTPDomain, handler)

	// configure TLS if certificates are provided
	if cfg.TLSCert != "" && cfg.TLSKey != "" {
		smtpServer = smtpServer.WithTLS(cfg.TLSCert, cfg.TLSKey)
	}

	// optional http server for health checks
	httpAddr := getEnvDefault("HTTP_ADDR", ":8080")
	httpSrv := newHTTPServer(httpAddr, logger)
	go func() {
		logger.Info("http server starting", "addr", httpAddr)
		if err := httpSrv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server error", "err", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start smtp server
	go func() {
		if err := smtpServer.Start(ctx); err != nil {
			logger.Fatal("smtp server error", "err", err)
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	logger.Info("shutting down. bye!")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = httpSrv.Shutdown(shutdownCtx)
}
