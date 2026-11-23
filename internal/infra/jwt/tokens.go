package jwtutil

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"boilerplate/internal/config"
	"boilerplate/internal/domain/entity"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
)

// GenerateAccessToken signs a JWT access token with all identity and permission claims.
// All data is baked in — downstream middleware never needs a DB lookup.
func GenerateAccessToken(claims TokenClaims, cfg config.JWTConfig) (string, error) {
	mapClaims := jwt.MapClaims{
		"sub": claims.ActorID,
		"iss": cfg.Issuer,
		"exp": time.Now().Add(cfg.AccessExpiry).Unix(),

		"actor_id":      claims.ActorID,
		"actor_type":    claims.ActorType,
		"email":         claims.Email,
		"display_name":  claims.DisplayName,
		"is_superadmin": claims.IsSuperAdmin,
		"status":        claims.Status,

		"can_read":   claims.CanRead,
		"can_write":  claims.CanWrite,
		"can_delete": claims.CanDelete,
		"can_manage": claims.CanManage,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mapClaims)
	return token.SignedString([]byte(cfg.JWTSecretKey))
}

// ValidateJWT parses and verifies a JWT token string.
// Returns the raw MapClaims on success for the auth middleware to extract.
func ValidateJWT(tokenString string, cfg config.JWTConfig) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString, claims,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.JWTSecretKey), nil
		},
	)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, jwt.ErrTokenSignatureInvalid
	}
	return claims, nil
}

// GenerateRefreshToken returns a cryptographically random opaque token.
// NOT a JWT — stored in the DB and exchanged for new access tokens.
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// MintTokenPair generates a matched access + refresh token pair.
// Does NOT persist anything — the caller saves the returned RefreshToken record.
func MintTokenPair(
	actorID string,
	actorType string,
	claims TokenClaims,
	cfg config.JWTConfig,
) (accessToken string, record *entity.RefreshToken, err error) {
	accessToken, err = GenerateAccessToken(claims, cfg)
	if err != nil {
		return "", nil, err
	}

	rawRefreshToken, err := GenerateRefreshToken()
	if err != nil {
		return "", nil, err
	}

	record = &entity.RefreshToken{
		ID:        ulid.Make().String(),
		ActorID:   actorID,
		ActorType: actorType,
		Token:     rawRefreshToken,
		ExpiresAt: time.Now().Add(cfg.RefreshExpiry),
	}

	return accessToken, record, nil
}

// CryptoRandRead is an exported wrapper around crypto/rand.Read for use outside this package.
func CryptoRandRead(b []byte) (int, error) {
	return rand.Read(b)
}
