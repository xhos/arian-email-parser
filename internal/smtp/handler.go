package smtp

import (
	"arian-parser/internal/api"
	"arian-parser/internal/domain"
	"arian-parser/internal/email"
	_ "arian-parser/internal/email/all"
	pb "arian-parser/internal/gen/arian/v1"
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

	if os.Getenv("SAVE_EML") != "" {
		if err := h.saveEmailToFile(userUUID, from, data); err != nil {
			h.Log.Warn("failed to save debug email file", "err", err)
		}
	}

	// resolve user id
	user, err := h.API.GetUser(userUUID)
	if err != nil {
		h.Log.Error("user not found", "user_uuid", userUUID, "err", err)
		return nil // accept email to avoid retries
	}
	userID := user.Id
	h.Log.Info("found user", "user_id", userID)

	msg, decoded, err := email.ParseMessage(data)
	if err != nil {
		h.Log.Error("failed to parse email message", "err", err)
		return nil
	}

	meta, err := parser.ToEmailMeta(fmt.Sprintf("%s-%d", userUUID, len(data)), msg, decoded)
	if err != nil {
		h.Log.Error("failed to parse email metadata", "err", err)
		return nil
	}

	h.Log.Debug("email content", "subject", meta.Subject, "text", meta.Text)

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

	h.Log.Debug("parsed transaction",
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
		accountMap[fmt.Sprintf("%s-%s", strings.ToLower(acc.Bank), acc.Name)] = int(acc.Id)
	}

	if err := h.resolveAccount(userUUID, txn, accountMap, user); err != nil {
		return err
	}

	if err := h.API.CreateTransaction(userID, txn); err != nil {
		h.Log.Error("failed to create transaction", "user_uuid", userUUID, "err", err)
		return nil
	}

	return nil
}

func (h *EmailHandler) saveEmailToFile(userUUID, from string, data []byte) error {
	const debugDir = "debug_emails"
	if err := os.MkdirAll(debugDir, 0755); err != nil {
		return fmt.Errorf("failed to create debug directory: %w", err)
	}

	msg, _, err := email.ParseMessage(data)
	if err != nil {
		h.Log.Warn("failed to parse email for debug file, saving raw", "err", err)
		timestamp := time.Now().Format("20060102-150405")
		filename := fmt.Sprintf("%s_%s_%s.eml", userUUID, timestamp, strings.ReplaceAll(from, "@", "_at_"))
		filePath := filepath.Join(debugDir, filename)

		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return fmt.Errorf("failed to write debug email file: %w", err)
		}
		h.Log.Info("saved raw debug email file", "path", filePath, "size", len(data))
		return nil
	}

	subject := msg.Header.Get("Subject")
	sanitizedSubject := sanitizeFilename(subject)
	timestamp := time.Now().Format("20060102-150405")

	filename := fmt.Sprintf("%s_%s_%s_%s.eml", userUUID, timestamp, sanitizedSubject, strings.ReplaceAll(from, "@", "_at_"))
	filePath := filepath.Join(debugDir, filename)

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write debug email file: %w", err)
	}

	h.Log.Info("saved debug email file", "path", filePath, "size", len(data), "subject", subject)
	return nil
}

func (h *EmailHandler) resolveAccount(userUUID string, txn *domain.Transaction, accountMap map[string]int, user *pb.User) error {
	noAccountParsed := txn.TxAccount == ""
	noDefaultAccount := user.GetDefaultAccountId() <= 0

	if noAccountParsed {
		if noDefaultAccount {
			h.Log.Warn("no account parsed and no default set; skipping", "user_uuid", userUUID)
			return nil
		}
		txn.AccountID = int(user.GetDefaultAccountId())
		return nil
	}

	cleanAccount := strings.TrimLeft(txn.TxAccount, "*")
	accountKey := fmt.Sprintf("%s-%s", strings.ToLower(txn.TxBank), cleanAccount)

	if existingAccountID, exists := accountMap[accountKey]; exists {
		txn.AccountID = existingAccountID
		return nil
	}

	account, err := h.API.CreateAccount(userUUID, cleanAccount, txn.TxBank)
	if err != nil {
		return fmt.Errorf("failed to create account for %s-%s: %w", txn.TxBank, cleanAccount, err)
	}

	txn.AccountID = int(account.Id)
	return nil
}

func sanitizeFilename(subject string) string {
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", " "}
	sanitized := subject
	for _, char := range invalid {
		sanitized = strings.ReplaceAll(sanitized, char, "_")
	}

	const maxFilenameLength = 50
	if len(sanitized) > maxFilenameLength {
		sanitized = sanitized[:maxFilenameLength]
	}

	sanitized = strings.TrimRight(sanitized, "_")

	if sanitized == "" {
		return "no-subject"
	}

	return sanitized
}
