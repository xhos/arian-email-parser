package config

import (
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

	TLSCert     string // TLS certificate file path
	TLSKey      string // TLS key file path
	TLSRequired bool   // enforce TLS for SMTP connections (default: true if certs provided)

	UnsafeSaveEML bool // save incoming emails to disk for debugging

	LogLevel log.Level // logging level
}

// safely parse whatever port or address the user provides
// handdles cases like "8080", ":8080", "127.0.0.1:8080"
func parseAddress(port string) string {
	port = strings.TrimSpace(port)
	if strings.Contains(port, ":") {
		return port
	}
	return ":" + port
}

func Load() Config {
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

	// SMTP and gRPC addresses with defaults
	smtpAddress := os.Getenv("SMTP_PORT")
	if smtpAddress == "" {
		smtpAddress = "127.0.0.1:2525"
	}

	grpcAddress := os.Getenv("GRPC_PORT")
	if grpcAddress == "" {
		grpcAddress = "127.0.0.1:50052"
	}

	logLevel, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		logLevel = log.InfoLevel
	}

	tlsCert := os.Getenv("TLS_CERT")
	tlsKey := os.Getenv("TLS_KEY")

	// TLS is required by default when certificates are provided
	// Can be disabled with UNSAFE_DISABLE_TLS_REQUIRED=true
	tlsRequired := tlsCert != "" && tlsKey != "" && os.Getenv("UNSAFE_DISABLE_TLS_REQUIRED") == ""

	return Config{
		NullCoreURL:   nullCoreURL,
		APIKey:        apiKey,
		Domain:        domain,
		SMTPAddress:   parseAddress(smtpAddress),
		GRPCAddress:   parseAddress(grpcAddress),
		TLSCert:       tlsCert,
		TLSKey:        tlsKey,
		TLSRequired:   tlsRequired,
		UnsafeSaveEML: os.Getenv("UNSAFE_SAVE_EML") != "",
		LogLevel:      logLevel,
	}
}
