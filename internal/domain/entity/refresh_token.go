package entity

import "time"

// RefreshToken is an opaque random token stored in the DB that lets actors
// obtain new access tokens without re-logging in.
// ActorType determines which table to look up when re-issuing access tokens.
type RefreshToken struct {
	ID        string    `gorm:"size:60;primaryKey"`
	ActorID   string    `gorm:"column:actor_id;size:60;not null;index"`   // User.ID | Admin.ID | Staff.ID
	ActorType string    `gorm:"column:actor_type;size:20;not null"`        // "user" | "admin" | "staff"
	Token     string    `gorm:"column:token;size:255;uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"column:expires_at;not null"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
