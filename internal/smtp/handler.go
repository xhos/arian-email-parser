package smtp

import (
	"arian-parser/internal/api"
	_ "arian-parser/internal/email/all"
	"arian-parser/internal/parser"
	"fmt"
	"strings"

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

	// validate user exists
	user, err := h.API.GetUser(userUUID)
	if err != nil {
		h.Log.Error("user not found", "user_uuid", userUUID, "err", err)
		return fmt.Errorf("user %s not found: %w", userUUID, err)
	}

	userID := user.Id
	h.Log.Info("found user", "user_uuid", userUUID, "user_id", userID)

	// parse email into metadata
	meta, err := parser.ToEmailMeta(fmt.Sprintf("%s-%d", userUUID, len(data)), data)
	if err != nil {
		h.Log.Error("failed to parse email metadata", "err", err)
		return err
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
