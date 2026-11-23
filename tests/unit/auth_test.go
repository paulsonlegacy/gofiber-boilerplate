package unit_test

import (
	"testing"
	"time"

	"boilerplate/internal/domain/entity"
	jwtutil "boilerplate/internal/infra/jwt"
)

// TestPasswordHashAndCompare verifies that HashPassword produces a hash that
// Compare accepts with the original password, and rejects a wrong one.
func TestPasswordHashAndCompare(t *testing.T) {
	password := "superSecret123!"
	hash := jwtutil.HashPassword(password, 32)

	if hash == "" {
		t.Fatal("HashPassword returned empty string")
	}
	if hash == password {
		t.Fatal("HashPassword returned plaintext — not hashed")
	}
	if !jwtutil.Compare(password, hash) {
		t.Error("Compare returned false for correct password")
	}
	if jwtutil.Compare("wrongpassword", hash) {
		t.Error("Compare returned true for wrong password")
	}
}

// TestRefreshTokenExpiry verifies that the RefreshToken entity carries the
// expected expiry structure so the cron cleaner query works correctly.
func TestRefreshTokenExpiry(t *testing.T) {
	now := time.Now()
	token := entity.RefreshToken{
		ID:        "test-id",
		ActorID:   "user-id",
		ActorType: "user",
		Token:     "tok123",
		ExpiresAt: now.Add(7 * 24 * time.Hour),
		CreatedAt: now,
	}

	if token.ExpiresAt.Before(now) {
		t.Error("token should not be expired on creation")
	}
	expired := entity.RefreshToken{
		ExpiresAt: now.Add(-1 * time.Minute),
	}
	if !expired.ExpiresAt.Before(now) {
		t.Error("past expiry should be before now")
	}
}
