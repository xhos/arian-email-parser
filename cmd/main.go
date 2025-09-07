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
	requiredEnvs := []string{"API_KEY", "ARIAND_URL", "DOMAIN"}
	for _, env := range requiredEnvs {
		if os.Getenv(env) == "" {
			panic(env + " environment variable is required")
		}
	}

	return Config{
		AriandURL: os.Getenv("ARIAND_URL"),
		APIKey:    os.Getenv("API_KEY"),
		SMTPAddr:  smtpAddr,
		Domain:    os.Getenv("DOMAIN"),
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

	// 2. initialize services
	handler := smtp.NewEmailHandler(apiClient, logger)
	smtpServer := smtp.NewServer(cfg.SMTPAddr, cfg.Domain, handler)
	if cfg.TLSCert != "" && cfg.TLSKey != "" {
		smtpServer = smtpServer.WithTLS(cfg.TLSCert, cfg.TLSKey)
	}

	grpcHealthSrv, err := grpc.NewHealthServer(cfg.GRPCAddr)
	if err != nil {
		logger.Fatal("grpc health server init", "err", err)
	}

	// 3. start servers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		logger.Info("grpc health server starting", "addr", cfg.GRPCAddr)
		if err := grpcHealthSrv.Start(); err != nil {
			logger.Error("grpc health server error", "err", err)
		}
	}()

	go func() {
		if err := smtpServer.Start(ctx); err != nil {
			logger.Fatal("smtp server error", "err", err)
		}
	}()

	// 4. graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	logger.Info("shutting down. bye!")

	cancel()
	grpcHealthSrv.Stop()
}
