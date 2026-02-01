package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"null-email-parser/internal/api"
	"null-email-parser/internal/config"
	"null-email-parser/internal/grpc"
	"null-email-parser/internal/smtp"
	"null-email-parser/internal/version"

	"github.com/charmbracelet/log"
)

func main() {
	cfg := config.Load()

	// ----- logger -----------------
	logger := log.NewWithOptions(os.Stdout, log.Options{
		Prefix: "email-parser",
		Level:  cfg.LogLevel,
	})

	logger.Info("starting email-parser", "version", version.FullVersion())

	logger.Debug("debug is enabled")

	if os.Getenv("SAVE_EML") != "" {
		logger.Info("SAVE_EML is enabled, emails will be saved to disk")
	}

	// ----- api client -------------
	apiClient, err := api.NewClient(cfg.NullCoreURL, "", cfg.APIKey)
	if err != nil {
		logger.Fatal("api client init", "err", err)
	}
	defer func() {
		if err := apiClient.Close(); err != nil {
			logger.Error("failed to close gRPC connection", "err", err)
		}
	}()

	// ----- connectivity check -----
	logger.Info("checking null-core connectivity", "url", cfg.NullCoreURL)
	if err := apiClient.Ping(); err != nil {
		logger.Fatal("null-core not reachable", "err", err)
	}
	logger.Info("null-core connectivity confirmed")

	// ----- services ---------------
	handler := smtp.NewEmailHandler(apiClient, logger)
	smtpServer := smtp.NewServer(cfg.SMTPAddress, cfg.Domain, handler)
	if cfg.TLSCert != "" && cfg.TLSKey != "" {
		smtpServer = smtpServer.WithTLS(cfg.TLSCert, cfg.TLSKey)
	}

	grpcHealthSrv, err := grpc.NewHealthServer(cfg.GRPCAddress)
	if err != nil {
		logger.Fatal("grpc health server init", "err", err)
	}

	// ----- servers ----------------
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		logger.Info("grpc health server starting", "address", cfg.GRPCAddress)
		if err := grpcHealthSrv.Start(); err != nil {
			logger.Error("grpc health server error", "err", err)
		}
	}()

	go func() {
		if err := smtpServer.Start(ctx); err != nil {
			logger.Fatal("smtp server error", "err", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	logger.Info("shutting down. bye!")

	cancel()
	grpcHealthSrv.Stop()
}
