package smtp

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/mhale/smtpd"
)

type Handler interface {
	ProcessEmail(userID string, from string, to []string, data []byte) error
}

type Server struct {
	addr       string
	domain     string
	handler    Handler
	log        *log.Logger
	smtpServer *smtpd.Server
	tlsCert    string
	tlsKey     string
}

func NewServer(addr, domain string, handler Handler) *Server {
	return &Server{
		addr:    addr,
		domain:  domain,
		handler: handler,
		log:     log.NewWithOptions(nil, log.Options{Prefix: "smtp"}),
	}
}

func (s *Server) WithTLS(certFile, keyFile string) *Server {
	s.tlsCert = certFile
	s.tlsKey = keyFile
	return s
}

func (s *Server) Start(ctx context.Context) error {
	s.smtpServer = &smtpd.Server{
		Addr:     s.addr,
		Handler:  s.mailHandler,
		Appname:  "arian-parser",
		Hostname: s.domain,
	}

	// configure TLS if certificates are provided
	if s.tlsCert != "" && s.tlsKey != "" {
		cert, err := tls.LoadX509KeyPair(s.tlsCert, s.tlsKey)
		if err != nil {
			return fmt.Errorf("failed to load TLS certificate: %w", err)
		}

		s.smtpServer.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			ServerName:   s.domain,
		}
		s.log.Info("starting smtp server with TLS", "addr", s.addr, "domain", s.domain)
	} else {
		s.log.Info("starting smtp server without TLS", "addr", s.addr, "domain", s.domain)
	}

	go func() {
		<-ctx.Done()
		s.log.Info("stopping smtp server")
		if err := s.smtpServer.Close(); err != nil {
			s.log.Error("error stopping smtp server", "err", err)
		}
	}()

	return s.smtpServer.ListenAndServe()
}

var uuidPattern = regexp.MustCompile(`^([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})@`)

func (s *Server) mailHandler(origin net.Addr, from string, to []string, data []byte) error {
	s.log.Debug("received email", "from", from, "to", to, "size", len(data))

	// extract uuid from first recipient
	if len(to) == 0 {
		return fmt.Errorf("no recipients")
	}

	recipient := strings.ToLower(to[0])

	// in debug mode, allow debug@domain.com emails
	if os.Getenv("DEBUG") != "" && strings.HasPrefix(recipient, "debug@") {
		s.log.Info("processing debug email", "from", from, "to", recipient)
		return s.handler.ProcessEmail("debug", from, to, data)
	}

	matches := uuidPattern.FindStringSubmatch(recipient)
	if len(matches) < 2 {
		s.log.Warn("invalid recipient format", "to", recipient)
		return fmt.Errorf("invalid recipient format, expected <uuid>@domain")
	}

	userID := matches[1]
	s.log.Info("processing email for user", "user_id", userID, "from", from)

	return s.handler.ProcessEmail(userID, from, to, data)
}
