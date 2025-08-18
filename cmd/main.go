package main

import (
	"context"
	"flag"
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
	AriandURL string
	APIKey    string
	SMTPAddr  string
	Domain    string
	TLSCert   string
	TLSKey    string
	GRPCAddr  string
}

func loadConfig(smtpAddr, grpcAddr string) Config {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		panic("API_KEY environment variable is required")
	}

	ariandURL := os.Getenv("ARIAND_URL")
	if ariandURL == "" {
		panic("ARIAND_URL environment variable is required")
	}

	domain := os.Getenv("DOMAIN")
	if domain == "" {
		panic("DOMAIN environment variable is required")
	}

	return Config{
		AriandURL: ariandURL,
		APIKey:    apiKey,
		SMTPAddr:  smtpAddr,
		Domain:    domain,
		TLSCert:   os.Getenv("TLS_CERT"),
		TLSKey:    os.Getenv("TLS_KEY"),
		GRPCAddr:  grpcAddr,
	}
}

func parseLogLevel(level string) log.Level {
	switch level {
	case "debug":
		return log.DebugLevel
	case "info":
		return log.InfoLevel
	case "warn":
		return log.WarnLevel
	case "error":
		return log.ErrorLevel
	default:
		return log.InfoLevel
	}
}

func main() {
	smtpPort := flag.String("smtp-port", "2525", "SMTP server port")
	grpcPort := flag.String("port", "50052", "gRPC health server port")
	flag.Parse()

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	logger := log.NewWithOptions(os.Stdout, log.Options{Prefix: "arian", Level: parseLogLevel(logLevel)})
	logger.Info("starting application",
		"app", version.RepoName,
		"version", version.Version(),
		"commit", version.GitCommit,
		"branch", version.GitBranch,
		"built", version.BuildTime,
		"repo", version.RepoURL)

	cfg := loadConfig(":"+*smtpPort, ":"+*grpcPort)

	// 1. initialize the API Client (skip in debug mode)
	var apiClient *api.Client
	if os.Getenv("DEBUG") == "" {
		var err error
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
	smtpServer := smtp.NewServer(cfg.SMTPAddr, cfg.Domain, handler)

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
