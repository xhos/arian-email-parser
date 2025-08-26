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

func (h *EmailHandler) ProcessEmail(userUUID, from string, to []string, data []byte) error {
	h.Log.Info("processing email", "user_uuid", userUUID, "from", from)

	debug := os.Getenv("DEBUG") != ""

	if debug {
		if err := h.saveEmailToFile(userUUID, from, data); err != nil {
			h.Log.Warn("failed to save debug email file", "err", err)
		}
	}

	// resolve user id (or debug default)
	userID := "1"
	if !debug {
		user, err := h.API.GetUser(userUUID)
		if err != nil {
			h.Log.Error("user not found", "user_uuid", userUUID, "err", err)
			return nil // accept email to avoid retries
		}
		userID = user.Id
		h.Log.Info("found user", "user_id", userID)
	} else {
		h.Log.Info("debug enabled: skipping user validation", "user_uuid", userUUID)
	}

	decoded, err := email.DecodeEmailContent(data)
	if err != nil {
		h.Log.Error("failed to decode email content", "err", err)
		return nil
	}

	h.Log.Debug("decoded email", "content", decoded)

	meta, err := parser.ToEmailMeta(fmt.Sprintf("%s-%d", userUUID, len(data)), decoded)
	if err != nil {
		h.Log.Error("failed to parse email metadata", "err", err)
		return nil
	}

	prsr := parser.Find(meta)
	if prsr == nil {
		h.Log.Info("no parser matched", "user_uuid", userUUID, "subject", meta.Subject)
		return nil
	}

	txn, err := prsr.Parse(meta)
	if err != nil {
		h.Log.Error("parse failed", "user_uuid", userUUID, "err", err)
		return nil
	}
	if txn == nil {
		return nil
	}

	if debug {
		h.Log.Info("parsed transaction (debug mode)",
			"user_uuid", userUUID,
			"email_id", txn.EmailID,
			"tx_date", txn.TxDate,
			"bank", txn.TxBank,
			"account", txn.TxAccount,
			"amount", txn.TxAmount,
			"currency", txn.TxCurrency,
			"direction", txn.TxDirection,
			"description", txn.TxDesc,
		)
		return nil
	}

	accounts, err := h.API.GetAccounts(userID)
	if err != nil {
		h.Log.Error("failed to fetch accounts", "err", err)
		return nil
	}

	accountMap := make(map[string]int, len(accounts))
	for _, acc := range accounts {
		if acc.Name == "" {
			continue
		}
		accountMap[fmt.Sprintf("%s-%s", acc.Bank, acc.Name)] = int(acc.Id)
	}

	cleanAccount := strings.TrimLeft(txn.TxAccount, "*")
	accountKey := fmt.Sprintf("%s-%s", txn.TxBank, cleanAccount)

	accountID, ok := accountMap[accountKey]
	if !ok {
		user, err := h.API.GetUser(userUUID)
		if err != nil {
			h.Log.Error("failed to get user for default account", "err", err)
			return nil
		}
		if user.GetDefaultAccountId() <= 0 {
			h.Log.Warn("unrecognized account and no default account set; skipping",
				"user_uuid", userUUID, "bank", txn.TxBank, "account", cleanAccount)
			return nil
		}
		accountID = int(user.GetDefaultAccountId())
		h.Log.Info("using default account for unrecognized account",
			"user_uuid", userUUID, "bank", txn.TxBank, "account", cleanAccount, "default_account_id", accountID)
	}

	txn.AccountID = accountID

	h.Log.Info("creating transaction",
		"user_uuid", userUUID, "account_id", accountID, "amount", txn.TxAmount, "description", txn.TxDesc)

	if err := h.API.CreateTransaction(userID, txn); err != nil {
		h.Log.Error("failed to create transaction", "user_uuid", userUUID, "err", err)
		return nil
	}

	h.Log.Info("successfully processed email", "user_uuid", userUUID, "email_id", txn.EmailID)
	return nil
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
