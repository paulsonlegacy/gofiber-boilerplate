package entity

import "time"

// User is the primary registered actor in the system.
// Rename to "Organization", "Company", "Account", or "Customer" as needed for your domain.
// For multi-tenant apps, this represents the top-level account owner (team, org, company).
// For consumer apps, this is the individual end user.
type User struct {
	ID        string      `gorm:"column:id;primaryKey;size:60"`
	Name      string      `gorm:"column:name;size:120;not null"`
	Email     string      `gorm:"column:email;size:255;uniqueIndex;not null"`
	Phone     string      `gorm:"column:phone;size:20;uniqueIndex;not null"`
	Password  string      `gorm:"column:password;size:255;not null"`
	Status    ActorStatus `gorm:"column:status;size:20;not null;default:pending"`

	// Password reset
	ResetToken          *string    `gorm:"column:reset_token;size:255;null"`
	ResetTokenExpiresAt *time.Time `gorm:"column:reset_token_expires_at;null"`

	// Optional profile fields — add or remove as needed
	AvatarURL *string `gorm:"column:avatar_url;size:512;null"`
	Bio       *string `gorm:"column:bio;size:1000;null"`

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`

	// Relations
	Wallet *Wallet `gorm:"foreignKey:UserID;references:ID"`
}

func (User) TableName() string {
	return "users"
}
