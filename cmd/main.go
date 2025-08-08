package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"arian-parser/internal/api"
	"arian-parser/internal/grpc"
	"arian-parser/internal/smtp"
	"arian-parser/internal/version"

	"github.com/charmbracelet/log"
)

type Config struct {
	AriandURL  string
	APIKey     string
	SMTPAddr   string
	SMTPDomain string
	TLSCert    string
	TLSKey     string
	GRPCAddr   string
}

func loadConfig() (Config, error) {
	cfg := Config{
		AriandURL:  os.Getenv("ARIAND_URL"),
		APIKey:     os.Getenv("API_KEY"),
		SMTPAddr:   getEnvDefault("SMTP_ADDR", ":2525"),
		SMTPDomain: getEnvDefault("SMTP_DOMAIN", "localhost"),
		TLSCert:    os.Getenv("TLS_CERT"),
		TLSKey:     os.Getenv("TLS_KEY"),
		GRPCAddr:   getEnvDefault("GRPC_ADDR", ":50052"),
	}

	// in debug mode, ARIAND_URL and API_KEY are optional
	if os.Getenv("DEBUG") == "" {
		if cfg.AriandURL == "" {
			return cfg, errors.New("ARIAND_URL must be set")
		}
		if cfg.APIKey == "" {
			return cfg, errors.New("API_KEY must be set")
		}
	}

	return cfg, nil
}

func getEnvDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	var showVersion = flag.Bool("version", false, "show version information")
	var showVersionFull = flag.Bool("version-full", false, "show full version information with build details")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.Version())
		return
	}
	if *showVersionFull {
		fmt.Println(version.FullVersion())
		return
	}

	logger := log.NewWithOptions(os.Stdout, log.Options{Prefix: "arian"})
	logger.Info("starting application",
		"app", version.RepoName,
		"version", version.Version(),
		"commit", version.GitCommit,
		"branch", version.GitBranch,
		"built", version.BuildTime,
		"repo", version.RepoURL)

	cfg, err := loadConfig()
	if err != nil {
		logger.Fatal("config error", "err", err)
	}

	// 1. initialize the API Client (skip in debug mode)
	var apiClient *api.Client
	if os.Getenv("DEBUG") == "" {
		apiClient, err = api.NewClient(cfg.AriandURL, "", cfg.APIKey)
		if err != nil {
			logger.Fatal("api client init", "err", err)
		}
		defer func() {
			if err := apiClient.Close(); err != nil {
				logger.Error("failed to close gRPC connection", "err", err)
			}
		}()
	} else {
		logger.Info("debug mode: skipping API client initialization")
	}

	// 2. initialize the email handler and smtp server
	handler := smtp.NewEmailHandler(apiClient, logger)
	smtpServer := smtp.NewServer(cfg.SMTPAddr, cfg.SMTPDomain, handler)

	// configure TLS if certificates are provided
	if cfg.TLSCert != "" && cfg.TLSKey != "" {
		smtpServer = smtpServer.WithTLS(cfg.TLSCert, cfg.TLSKey)
	}

	// gRPC health server
	grpcHealthSrv, err := grpc.NewHealthServer(cfg.GRPCAddr)
	if err != nil {
		logger.Fatal("grpc health server init", "err", err)
	}
	go func() {
		logger.Info("grpc health server starting", "addr", cfg.GRPCAddr)
		if err := grpcHealthSrv.Start(); err != nil {
			logger.Error("grpc health server error", "err", err)
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

	grpcHealthSrv.Stop()
}
