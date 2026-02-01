package config

import (
	"flag"
	"os"
	"strings"

	"github.com/charmbracelet/log"
)

type Config struct {
	NullCoreURL string // null-core service URL
	APIKey      string // internal API key for authenticating requests
	Domain      string // domain for the SMTP server

	SMTPAddress string // SMTP server address
	GRPCAddress string // gRPC server address

	TLSCert string // TLS certificate file path
	TLSKey  string // TLS key file path

	LogLevel log.Level // logging level
}

// parseAddress ensures the address is in the correct format for network listeners.
// If the input is just a port (e.g. "2525"), it returns ":2525".
// If the input is already an address (e.g. "0.0.0.0:2525" or ":2525"), it returns it unchanged.
// Examples:
//
//	parseAddress("2525")         // ":2525"
//	parseAddress(":2525")        // ":2525"
//	parseAddress("0.0.0.0:2525") // "0.0.0.0:2525"
func parseAddress(port string) string {
	port = strings.TrimSpace(port)
	if strings.Contains(port, ":") {
		return port
	}
	return ":" + port
}

// Load reads configuration from environment variables and command-line flags
func Load() Config {
	smtpAddress := flag.String("smtp-port", "2525", "SMTP server port (e.g. 2525, :2525, 0.0.0.0:2525)")
	grpcAddress := flag.String("grpc-port", "50052", "gRPC server port (e.g. 50052, :50052, 0.0.0.0:50052)")

	flag.Parse()

	nullCoreURL := os.Getenv("NULL_CORE_URL")
	if nullCoreURL == "" {
		panic("NULL_CORE_URL environment variable is required")
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		panic("API_KEY environment variable is required")
	}

	domain := os.Getenv("DOMAIN")
	if domain == "" {
		panic("DOMAIN environment variable is required")
	}

	logLevel, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		logLevel = log.InfoLevel
	}

	return Config{
		NullCoreURL: nullCoreURL,
		APIKey:      apiKey,
		Domain:      domain,
		SMTPAddress: parseAddress(*smtpAddress),
		GRPCAddress: parseAddress(*grpcAddress),
		TLSCert:     os.Getenv("TLS_CERT"),
		TLSKey:      os.Getenv("TLS_KEY"),
		LogLevel:    logLevel,
	}
}
