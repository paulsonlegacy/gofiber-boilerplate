package mail

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/smtp"
	"path/filepath"
	"runtime"

	"boilerplate/internal/config"
)

// SMTPMailer sends emails via a standard SMTP server.
// Compatible with Gmail (with App Password), Mailgun SMTP, AWS SES SMTP, etc.
type SMTPMailer struct {
	cfg config.SMTPProviderConfig
}

// NewSMTPMailer creates a new SMTPMailer from the given provider config.
func NewSMTPMailer(cfg config.SMTPProviderConfig) *SMTPMailer {
	return &SMTPMailer{cfg: cfg}
}

// Send compiles the HTML template and delivers the email via SMTP.
func (m *SMTPMailer) Send(ctx context.Context, email Email) error {
	if len(email.To) == 0 {
		return fmt.Errorf("no recipients specified")
	}

	body, err := m.renderTemplate(email.Template, email.Data)
	if err != nil {
		return fmt.Errorf("template render failed: %w", err)
	}

	msg := m.buildMessage(email, body)
	return m.sendSMTP(email.To, msg)
}

func (m *SMTPMailer) renderTemplate(templateName string, data map[string]any) (string, error) {
	_, thisFile, _, _ := runtime.Caller(0)
	// Walk up to project root and locate templates/
	root := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..", "templates")
	path := filepath.Join(root, templateName)

	tmpl, err := template.ParseFiles(path)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (m *SMTPMailer) buildMessage(email Email, htmlBody string) []byte {
	from := m.cfg.From
	if from == "" {
		from = m.cfg.Username
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("From: %s\r\n", from))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", email.To[0]))
	if m.cfg.ReplyTo != "" {
		buf.WriteString(fmt.Sprintf("Reply-To: %s\r\n", m.cfg.ReplyTo))
	}
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", email.Subject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(htmlBody)
	return buf.Bytes()
}

func (m *SMTPMailer) sendSMTP(to []string, msg []byte) error {
	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)
	auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)

	tlsConfig := &tls.Config{
		ServerName: m.cfg.Host,
	}

	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer client.Quit()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}
	}

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}
	if err := client.Mail(m.cfg.Username); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	for _, r := range to {
		if err := client.Rcpt(r); err != nil {
			return fmt.Errorf("smtp rcpt to %s: %w", r, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	return w.Close()
}
