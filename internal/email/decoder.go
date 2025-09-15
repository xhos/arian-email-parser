package email

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"
)

// ParseMessage parses raw email data and returns both the message and plain text content
func ParseMessage(raw []byte) (*mail.Message, string, error) {
	msg, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return nil, "", fmt.Errorf("parsing email message: %w", err)
	}

	content, err := extractTextContent(msg)
	if err != nil {
		return nil, "", fmt.Errorf("extracting text content: %w", err)
	}

	return msg, content, nil
}

// extractTextContent extracts plain text content from a mail message
func extractTextContent(msg *mail.Message) (string, error) {
	ct := msg.Header.Get("Content-Type")
	if ct == "" {
		// Simple text email, read body directly
		body, err := io.ReadAll(msg.Body)
		if err != nil {
			return "", fmt.Errorf("reading simple email body: %w", err)
		}
		return string(body), nil
	}

	mediaType, params, err := mime.ParseMediaType(ct)
	if err != nil {
		return "", fmt.Errorf("parsing Content-Type: %w", err)
	}

	// Handle multipart emails
	if strings.HasPrefix(mediaType, "multipart/") {
		return extractMultipartText(msg.Body, params["boundary"])
	}

	// Handle single-part email with encoding
	encoding := msg.Header.Get("Content-Transfer-Encoding")
	var reader io.Reader = msg.Body
	if strings.EqualFold(encoding, "base64") {
		reader = base64.NewDecoder(base64.StdEncoding, msg.Body)
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("reading encoded body: %w", err)
	}

	return string(body), nil
}

// extractMultipartText walks through multipart email to find text content
func extractMultipartText(body io.Reader, boundary string) (string, error) {
	mr := multipart.NewReader(body, boundary)
	var htmlContent string

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("reading multipart: %w", err)
		}

		partType := part.Header.Get("Content-Type")

		// Prefer text/plain if available
		if strings.HasPrefix(partType, "text/plain") {
			return decodePart(part)
		}

		// Store HTML content as fallback
		if strings.HasPrefix(partType, "text/html") {
			if content, err := decodePart(part); err == nil {
				htmlContent = content
			}
		}

		// Handle nested multipart
		if strings.HasPrefix(partType, "multipart/") {
			_, params, err := mime.ParseMediaType(partType)
			if err != nil {
				continue
			}
			if nestedBoundary := params["boundary"]; nestedBoundary != "" {
				if text, err := extractMultipartText(part, nestedBoundary); err == nil && text != "" {
					return text, nil
				}
			}
		}
	}

	// If we have HTML content but no text/plain, use it
	if htmlContent != "" {
		return htmlContent, nil
	}

	return "", fmt.Errorf("no text/plain or text/html part found")
}

// decodePart reads and decodes a multipart section
func decodePart(part *multipart.Part) (string, error) {
	encoding := part.Header.Get("Content-Transfer-Encoding")
	var reader io.Reader = part

	if strings.EqualFold(encoding, "base64") {
		reader = base64.NewDecoder(base64.StdEncoding, part)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("decoding part: %w", err)
	}

	return string(data), nil
}

// DecodeEmailContent extracts plain text content from raw email data (legacy compatibility)
func DecodeEmailContent(raw []byte) (string, error) {
	_, content, err := ParseMessage(raw)
	return content, err
}
