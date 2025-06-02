package mailpit

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// UnreadIDs retrieves the IDs of all unread messages from Mailpit
func (c *Client) UnreadIDs() ([]string, error) {
	var resp struct {
		Messages []struct {
			ID   json.RawMessage `json:"ID"`
			Read json.RawMessage `json:"Read"`
		}
	}

	data, err := c.doRequest(http.MethodGet, "api/v1/messages")
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	var ids []string
	for _, m := range resp.Messages {
		var id string
		if err := json.Unmarshal(m.ID, &id); err != nil {
			c.log.Warn("failed to unmarshal message ID during GetUnreadEmails", "error", err)
			continue
		}
		var read bool
		if err := json.Unmarshal(m.Read, &read); err != nil {
			c.log.Warn("failed to unmarshal read status for message", "id", id, "error", err)
			continue
		}
		if !read {
			ids = append(ids, id)
		}
	}
	c.log.Info("unread messages", "count", len(ids))
	return ids, nil
}

// Message retrieves the full content of a message by its ID
func (c *Client) Message(id string) ([]byte, error) {
	if id == "" {
		return nil, fmt.Errorf("message ID cannot be empty")
	}
	return c.doRequest(http.MethodGet, fmt.Sprintf("api/v1/message/%s", id))
}
