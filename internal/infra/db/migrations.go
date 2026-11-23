package db

import (
	"boilerplate/internal/domain/entity"

	"gorm.io/gorm"
)

// AutoMigrateAll runs GORM auto-migration for all registered entities.
// Add new entities here as your domain grows. Order matters for foreign key constraints —
// parent tables must be listed before their dependents.
func AutoMigrateAll(db *gorm.DB) error {
	return db.AutoMigrate(
		// Core actors
		&entity.Admin{},
		&entity.User{},

		// Session management
		&entity.RefreshToken{},

		// Finance
		&entity.Wallet{},
		&entity.Transaction{},

		// Notifications
		&entity.Notification{},

		// Add your domain entities here:
		// &entity.Order{},
		// &entity.Subscription{},
		// &entity.Product{},
	)
}
