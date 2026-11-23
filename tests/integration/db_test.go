package integration_test

import (
	"context"
	"testing"

	"boilerplate/internal/domain/entity"
	database "boilerplate/internal/infra/db"

	"github.com/oklog/ulid/v2"
)

// TestSQLiteAutoMigrate verifies that AutoMigrateAll runs without error
// on a fresh in-memory SQLite database and that basic CRUD works.
func TestSQLiteAutoMigrate(t *testing.T) {
	db := database.NewSQLiteDB(":memory:")
	if db == nil {
		t.Fatal("NewSQLiteDB returned nil")
	}

	ctx := context.Background()

	// Insert a user
	user := &entity.User{
		ID:       ulid.Make().String(),
		Name:     "Test User",
		Email:    "test@example.com",
		Phone:    "08000000000",
		Password: "hashedpassword",
		Status:   entity.ActorStatusActive,
	}
	if err := db.WithContext(ctx).Create(user).Error; err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	// Fetch it back
	var found entity.User
	if err := db.WithContext(ctx).Where("email = ?", user.Email).First(&found).Error; err != nil {
		t.Fatalf("failed to fetch user: %v", err)
	}

	if found.Name != user.Name {
		t.Errorf("expected name %q, got %q", user.Name, found.Name)
	}

	// Delete it
	if err := db.WithContext(ctx).Delete(&entity.User{}, "id = ?", user.ID).Error; err != nil {
		t.Fatalf("failed to delete user: %v", err)
	}
}
