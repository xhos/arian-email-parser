package smtp

import (
	"arian-parser/internal/api"
	"arian-parser/internal/email"
	_ "arian-parser/internal/email/all"
	"arian-parser/internal/parser"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

// EmailHandler processes incoming SMTP emails
type EmailHandler struct {
	API *api.Client
	Log *log.Logger
}

func NewEmailHandler(apiClient *api.Client, log *log.Logger) *EmailHandler {
	return &EmailHandler{
		API: apiClient,
		Log: log.WithPrefix("handler"),
	}
}

// ProcessEmail implements the Handler interface
func (h *EmailHandler) ProcessEmail(userUUID string, from string, to []string, data []byte) error {
	h.Log.Info("processing email", "user_uuid", userUUID, "from", from)

	// save email to file if DEBUG is enabled
	if os.Getenv("DEBUG") != "" {
		if err := h.saveEmailToFile(userUUID, from, data); err != nil {
			h.Log.Warn("failed to save debug email file", "err", err)
		}
		// Log first 200 characters of raw email for debugging
		preview := string(data)
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		h.Log.Info("raw email preview", "preview", preview)
	}

	var userID string = "1" // default for debug mode

	// skip ARIAND connection in debug mode
	if os.Getenv("DEBUG") == "" {
		// validate user exists
		user, err := h.API.GetUser(userUUID)
		if err != nil {
			h.Log.Error("user not found", "user_uuid", userUUID, "err", err)
			return fmt.Errorf("user %s not found: %w", userUUID, err)
		}
		userID = user.Id
		h.Log.Info("found user", "user_uuid", userUUID, "user_id", userID)
	} else {
		h.Log.Info("debug mode: skipping user validation", "user_uuid", userUUID)
	}

	// decode raw email content
	decodedContent, err := email.DecodeEmailContent(data)
	if err != nil {
		h.Log.Error("failed to decode email content", "err", err)
		// Return nil to accept the email and prevent retries - this is our parsing issue, not sender's
		return nil
	}

	// save decoded content if DEBUG is enabled
	if os.Getenv("DEBUG") != "" {
		h.Log.Info("decoded email content", "content", decodedContent)
	}

	// parse email into metadata
	meta, err := parser.ToEmailMeta(fmt.Sprintf("%s-%d", userUUID, len(data)), decodedContent)
	if err != nil {
		h.Log.Error("failed to parse email metadata", "err", err)
		// Return nil to accept the email and prevent retries - this is our parsing issue, not sender's
		return nil
	}

	// find appropriate parser
	prsr := parser.Find(meta)
	if prsr == nil {
		h.Log.Info("no parser matched", "user_uuid", userUUID, "subject", meta.Subject)
		return nil
	}

	// parse transaction
	txn, err := prsr.Parse(meta)
	if err != nil {
		h.Log.Error("parse failed", "user_uuid", userUUID, "err", err)
		return err
	}
	if txn == nil {
		return nil
	}

	// in debug mode, just log the parsed transaction
	if os.Getenv("DEBUG") != "" {
		h.Log.Info("parsed transaction (debug mode)",
			"user_uuid", userUUID,
			"email_id", txn.EmailID,
			"tx_date", txn.TxDate,
			"bank", txn.TxBank,
			"account", txn.TxAccount,
			"amount", txn.TxAmount,
			"currency", txn.TxCurrency,
			"direction", txn.TxDirection,
			"description", txn.TxDesc)
		return nil
	}

	// get user's accounts
	accounts, err := h.API.GetAccounts(userID)
	if err != nil {
		h.Log.Error("failed to fetch accounts", "err", err)
		return err
	}

	// build account lookup map
	accountMap := make(map[string]int)
	for _, acc := range accounts {
		if acc.Name == "" {
			continue
		}
		key := fmt.Sprintf("%s-%s", acc.Bank, acc.Name)
		accountMap[key] = int(acc.Id)
	}

	// look up account
	cleanAccount := strings.TrimLeft(txn.TxAccount, "*")
	accountKey := fmt.Sprintf("%s-%s", txn.TxBank, cleanAccount)

	accountID, ok := accountMap[accountKey]
	if !ok {
		h.Log.Warn("unrecognized account; skipping",
			"user_uuid", userUUID,
			"bank", txn.TxBank,
			"account", cleanAccount)
		return nil
	}

	txn.AccountID = accountID

	h.Log.Info("creating transaction",
		"user_uuid", userUUID,
		"account_id", accountID,
		"amount", txn.TxAmount,
		"description", txn.TxDesc)

	return h.API.CreateTransaction(userID, txn)
}

// saveEmailToFile saves email content to debug directory when DEBUG env var is set
func (h *EmailHandler) saveEmailToFile(userUUID, from string, data []byte) error {
	debugDir := "debug_emails"
	if err := os.MkdirAll(debugDir, 0755); err != nil {
		return fmt.Errorf("failed to create debug directory: %w", err)
	}

	decodedContent, err := email.DecodeEmailContent(data)
	if err != nil {
		h.Log.Warn("failed to decode email for debug file, saving raw", "err", err)
		timestamp := time.Now().Format("20060102-150405")
		filename := fmt.Sprintf("%s_%s_%s.eml", userUUID, timestamp, strings.ReplaceAll(from, "@", "_at_"))
		filePath := filepath.Join(debugDir, filename)

		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return fmt.Errorf("failed to write debug email file: %w", err)
		}
		h.Log.Info("saved raw debug email file", "path", filePath, "size", len(data))
		return nil
	}

	subject := extractSubjectFromContent(decodedContent)
	sanitizedSubject := sanitizeFilename(subject)

	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s_%s_%s_%s.txt", userUUID, timestamp, sanitizedSubject, strings.ReplaceAll(from, "@", "_at_"))
	filePath := filepath.Join(debugDir, filename)

	if err := os.WriteFile(filePath, []byte(decodedContent), 0644); err != nil {
		return fmt.Errorf("failed to write debug email file: %w", err)
	}

	h.Log.Info("saved decoded debug email file", "path", filePath, "size", len(decodedContent), "subject", subject)
	return nil
}

// extractSubjectFromContent extracts subject line from decoded email content
func extractSubjectFromContent(content string) string {
	for line := range strings.SplitSeq(content, "\n") {
		if subject, found := strings.CutPrefix(line, "Subject:"); found {
			return strings.TrimSpace(subject)
		}
	}
	return "no-subject"
}

// sanitizeFilename removes invalid characters from subject for use in filename
func sanitizeFilename(subject string) string {
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", " "}
	sanitized := subject
	for _, char := range invalid {
		sanitized = strings.ReplaceAll(sanitized, char, "_")
	}

	if len(sanitized) > 50 {
		sanitized = sanitized[:50]
	}

	sanitized = strings.TrimRight(sanitized, "_")

	if sanitized == "" {
		return "no-subject"
	}

	return sanitized
}
