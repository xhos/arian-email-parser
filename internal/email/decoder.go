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

// DecodeEmailContent extracts plain text content from raw email data
func DecodeEmailContent(raw []byte) (string, error) {
	// Parse the email message
	msg, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return "", fmt.Errorf("parsing email message: %w", err)
	}

	// Get Content-Type header
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

// extractMultipartText walks through multipart email to find text/plain content
func extractMultipartText(body io.Reader, boundary string) (string, error) {
	mr := multipart.NewReader(body, boundary)

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("reading multipart: %w", err)
		}

		partType := part.Header.Get("Content-Type")
		if strings.HasPrefix(partType, "text/plain") {
			return decodePart(part)
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

	return "", fmt.Errorf("no text/plain part found")
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

	text := string(data)

	// Clean up forwarded email format (remove "> " prefixes and extract forwarded content)
	return cleanForwardedEmail(text), nil
}

// cleanForwardedEmail removes email forwarding artifacts and extracts the original content
func cleanForwardedEmail(text string) string {
	var lines []string

	// Remove "> " prefixes from quoted content
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimPrefix(line, "> ")
		if strings.TrimSpace(trimmed) != "" {
			lines = append(lines, trimmed)
		}
	}

	clean := strings.Join(lines, "\n")

	// Look for forwarded message marker and extract content after it
	markers := []string{
		"Forwarded Message",
		"---------- Forwarded message",
		"Begin forwarded message:",
	}

	for _, marker := range markers {
		if idx := strings.Index(clean, marker); idx != -1 {
			clean = clean[idx:]
			break
		}
	}

	return clean
}
