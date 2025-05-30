package email

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// this actully also marks as read the emails it gets, nice
func (c *Client) GetUnreadEmails(ctx context.Context) ([]string, error) {
	var response struct {
		Messages []struct{ ID, Read json.RawMessage }
	}

	data, err := c.doRequest(ctx, http.MethodGet, "api/v1/messages")
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &response); err != nil {
		c.log.Error("unmarshal failed", "error", err)
		return nil, err
	}

	var ids []string
	for _, m := range response.Messages {
		var id string
		json.Unmarshal(m.ID, &id)
		var read bool
		json.Unmarshal(m.Read, &read)
		if !read {
			ids = append(ids, id)
		}
	}

	c.log.Debug("unread messages", "count", len(ids))

	return ids, nil
}

func (c *Client) GetEmail(ctx context.Context, id string) ([]byte, error) {
	if id == "" {
		c.log.Error("empty message ID")
		return nil, errors.New("message ID cannot be empty")
	}

	data, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("api/v1/message/%s", id))
	if err != nil {
		c.log.Error("fetch content failed", "id", id, "error", err)
	}

	return data, err
}
