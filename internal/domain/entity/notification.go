package entity

import "time"

// Notification is an in-app alert for any actor.
// IsRead tracks whether the actor has seen it; optionally queue an email via the service layer.
type Notification struct {
	ID        string    `gorm:"column:id;primaryKey;size:60"`
	ActorID   string    `gorm:"column:actor_id;size:60;not null;index"`
	ActorType string    `gorm:"column:actor_type;size:20;not null"` // "user" | "admin" | "staff"
	Title     string    `gorm:"column:title;size:255;not null"`
	Message   string    `gorm:"column:message;size:1000;not null"`
	IsRead    bool      `gorm:"column:is_read;default:false"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;index"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Notification) TableName() string {
	return "notifications"
}
