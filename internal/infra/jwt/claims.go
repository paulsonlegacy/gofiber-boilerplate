// Package jwtutil provides JWT generation, validation, and password hashing utilities.
package jwtutil

import "github.com/golang-jwt/jwt/v5"

// TokenClaims holds all identity and permission data baked into a JWT access token.
// Permissions are cached here at login time so downstream middleware needs zero DB queries.
type TokenClaims struct {
	// Identity
	ActorID     string // User.ID | Admin.ID | Staff.ID
	ActorType   string // "user" | "admin" | "staff"
	Email       string
	DisplayName string
	IsSuperAdmin bool
	Status      string // "active" | "pending" | "suspended"

	// Permissions — for account owners (user/admin): all true.
	// For staff: sourced from StaffPermission table at login, then token-cached.
	CanRead   bool
	CanWrite  bool
	CanDelete bool
	CanManage bool // admin/owner-level operations within an account
}

// ClaimBool safely reads a boolean claim. Returns false on type mismatch or missing key.
func ClaimBool(claims jwt.MapClaims, key string) bool {
	v, ok := claims[key].(bool)
	return ok && v
}

// ClaimStr safely reads a string claim. Returns empty string on type mismatch or missing key.
func ClaimStr(claims jwt.MapClaims, key string) string {
	v, _ := claims[key].(string)
	return v
}
