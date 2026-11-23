// Package telegram provides structured alert delivery via the Telegram Bot API.
// Used for operational alerts — errors, critical events, system state changes.
// All methods are no-ops when Token or ChatID are empty, so it's safe to
// initialize without credentials in development.
package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// TelegramService sends formatted messages to a Telegram group or channel.
type TelegramService struct {
	token  string
	chatID string
	client *http.Client
}

// NewTelegramService creates a TelegramService.
// If token or chatID is empty, all Send calls are silent no-ops.
func NewTelegramService(token, chatID string) *TelegramService {
	return &TelegramService{
		token:  token,
		chatID: chatID,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Success sends a green-flagged alert.
func (t *TelegramService) Success(title, body string) error {
	return t.send(fmt.Sprintf("✅ *%s*\n\n%s", title, body))
}

// Error sends a red-flagged alert. Use for failures that require human attention.
func (t *TelegramService) Error(title, body string) error {
	return t.send(fmt.Sprintf("🚨 *%s*\n\n%s", title, body))
}

// Info sends a neutral informational alert.
func (t *TelegramService) Info(title, body string) error {
	return t.send(fmt.Sprintf("ℹ️ *%s*\n\n%s", title, body))
}

func (t *TelegramService) send(message string) error {
	if t == nil || t.token == "" || t.chatID == "" {
		return nil // not configured — silent no-op
	}

	payload := map[string]string{
		"chat_id":    t.chatID,
		"text":       message,
		"parse_mode": "Markdown",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("telegram: marshal: %w", err)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.token)
	resp, err := t.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("telegram: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram: unexpected status %d", resp.StatusCode)
	}
	return nil
}
