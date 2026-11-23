// Package mail provides a multi-provider email abstraction.
// All mail sending goes through the Mailer interface — adding a new provider
// requires only implementing the interface and adding it to the provider list.
package mail

import "context"

// Email is the universal email payload.
type Email struct {
	To       []string       // recipient addresses
	Subject  string
	Template string         // filename in the templates/ directory, e.g. "welcome.html"
	Data     map[string]any // template variables injected at render time
}

// Mailer is the interface all email providers must satisfy.
type Mailer interface {
	Send(ctx context.Context, email Email) error
}

// MultiProviderMailer tries each registered provider in order until one succeeds.
// This gives automatic fallback — if the primary SMTP server is down, the backup fires.
type MultiProviderMailer struct {
	Providers []Mailer
}

// NewMultiProviderMailer creates a MultiProviderMailer with the given providers in priority order.
func NewMultiProviderMailer(providers ...Mailer) *MultiProviderMailer {
	return &MultiProviderMailer{Providers: providers}
}

// Send tries each provider in order. Returns the last error if all providers fail.
func (m *MultiProviderMailer) Send(ctx context.Context, email Email) error {
	var lastErr error
	for _, provider := range m.Providers {
		if err := provider.Send(ctx, email); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}
	return lastErr
}
