package entity

import "time"

// Admin represents a platform operator.
// SuperAdmin admins have elevated privileges (e.g. creating other admins, hard deletes).
type Admin struct {
	ID           string      `gorm:"column:id;primaryKey;size:60"`
	FirstName    string      `gorm:"column:first_name;size:80;not null"`
	LastName     string      `gorm:"column:last_name;size:80;not null"`
	Email        string      `gorm:"column:email;size:255;uniqueIndex;not null"`
	Password     string      `gorm:"column:password;size:255;not null"`
	Status       ActorStatus `gorm:"column:status;size:20;not null;default:active"`
	IsSuperAdmin bool        `gorm:"column:is_superadmin;default:false"`

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Admin) TableName() string {
	return "admins"
}
