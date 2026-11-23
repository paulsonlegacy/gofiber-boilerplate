package queue

import "boilerplate/internal/infra/mail"

// ─── Job types ────────────────────────────────────────────────────────────────

type JobType string

const (
	TypeMail         JobType = "mail"
	TypeNotification JobType = "notification"
	// Add new job types here as your domain grows:
	// TypeExport JobType = "export"
	// TypeWebhook JobType = "webhook"
)

// ─── Job ─────────────────────────────────────────────────────────────────────

// Job is the unit of work placed on the queue.
// Type determines which worker branch handles it; Data carries the payload.
type Job struct {
	Type JobType
	Data any
}

// ─── Job data types ───────────────────────────────────────────────────────────

type MailJobData struct {
	Email *mail.Email
}

type NotificationJobData struct {
	ActorID   string
	ActorType string
	Title     string
	Message   string
	SendEmail bool
}

// ─── Job factories ────────────────────────────────────────────────────────────

// NewMailJob creates a fire-and-forget email delivery job.
func NewMailJob(email *mail.Email) Job {
	return Job{Type: TypeMail, Data: MailJobData{Email: email}}
}

// NewNotificationJob creates a notification persistence + optional email job.
func NewNotificationJob(actorID, actorType, title, message string, sendEmail bool) Job {
	return Job{Type: TypeNotification, Data: NotificationJobData{
		ActorID:   actorID,
		ActorType: actorType,
		Title:     title,
		Message:   message,
		SendEmail: sendEmail,
	}}
}
