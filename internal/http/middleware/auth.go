// Package middleware provides Fiber middleware for authentication, authorization,
// input sanitization, and maintenance mode.
package middleware

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"boilerplate/internal/config"
	"boilerplate/internal/http/apperror"
	"boilerplate/internal/http/response"
	jwtutil "boilerplate/internal/infra/jwt"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AuthMiddleware validates incoming JWT tokens and populates c.Locals with
// actor identity and permission flags. Downstream handlers read from Locals —
// zero DB queries are made per request after the auth middleware runs.
type AuthMiddleware struct {
	config.ServerConfig
	config.JWTConfig
	DB *gorm.DB
}

func NewAuthMiddleware(serverCfg config.ServerConfig, jwtCfg config.JWTConfig, db *gorm.DB) *AuthMiddleware {
	return &AuthMiddleware{serverCfg, jwtCfg, db}
}

// unprotectedRoutes lists paths that bypass JWT authentication entirely.
// Add any public endpoint here — register, login, webhook callbacks, etc.
var unprotectedRoutes = []*regexp.Regexp{
	regexp.MustCompile("^/api/auth/register$"),
	regexp.MustCompile("^/api/auth/login$"),
	regexp.MustCompile("^/api/auth/refresh-token$"),
	regexp.MustCompile("^/api/auth/forgot-password$"),
	regexp.MustCompile("^/api/auth/reset-password$"),
	regexp.MustCompile("^/health$"),
	regexp.MustCompile("^/metrics$"),
}

// Handler returns the JWT authentication middleware function.
func (m *AuthMiddleware) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {

		// Skip auth for public routes
		for _, route := range unprotectedRoutes {
			if route.MatchString(c.OriginalURL()) {
				return c.Next()
			}
		}

		// Extract token from Authorization header or ?token= query param (for WebSocket)
		var token string
		if authHeader := c.Get("Authorization"); authHeader != "" {
			extracted, err := extractTokenFromHeader(authHeader)
			if err != nil {
				return response.Unauthorized(c, apperror.NewUnauthorizedErr("Invalid authorization header format"))
			}
			token = extracted
		}
		if token == "" {
			if qt := c.Query("token"); qt != "" {
				token = strings.TrimSpace(strings.TrimPrefix(qt, "Bearer "))
			}
		}
		if token == "" {
			return response.Unauthorized(c, apperror.NewUnauthorizedErr("Missing authentication token"))
		}

		// Validate JWT signature
		claims, err := jwtutil.ValidateJWT(token, m.JWTConfig)
		if err != nil {
			zap.L().Debug("JWT validation failed",
				zap.String("path", c.Path()),
				zap.String("ip", c.IP()),
				zap.Error(err),
			)
			return response.Unauthorized(c, apperror.New(
				apperror.ErrCodeTokenInvalid, "Invalid token",
				fmt.Sprintf("token validation failed: %s", err.Error()),
			))
		}

		// Check expiry
		expFloat, ok := claims["exp"].(float64)
		if !ok {
			return response.Unauthorized(c, apperror.New(apperror.ErrCodeTokenInvalid, "Invalid token", "missing exp claim"))
		}
		if time.Now().After(time.Unix(int64(expFloat), 0)) {
			return response.Unauthorized(c, apperror.New(apperror.ErrCodeTokenExpired, "Token expired", "access token has expired"))
		}

		// Read identity claims
		actorID, _ := claims["actor_id"].(string)
		actorType, _ := claims["actor_type"].(string)
		status, _ := claims["status"].(string)

		if actorID == "" || actorType == "" {
			return response.Unauthorized(c, apperror.NewUnauthorizedErr("Token is missing identity claims"))
		}

		if status == "deleted" {
			return response.Unauthorized(c, apperror.New(apperror.ErrCodeAccountDeleted, "Account deleted", ""))
		}

		allowedStatuses := []string{"active", "approved", "pending"}
		isSuperAdmin, _ := claims["is_superadmin"].(bool)
		if !isSuperAdmin && !slices.Contains(allowedStatuses, status) {
			return response.Unauthorized(c, apperror.NewUnauthorizedErr("Permission denied"))
		}

		// Populate Locals — all downstream middleware and handlers read from here
		c.Locals("actor_id", actorID)
		c.Locals("actor_type", actorType)
		c.Locals("email", jwtutil.ClaimStr(claims, "email"))
		c.Locals("display_name", jwtutil.ClaimStr(claims, "display_name"))
		c.Locals("is_superadmin", isSuperAdmin)
		c.Locals("status", status)

		// Permission flags — default false if absent or malformed
		c.Locals("can_read", jwtutil.ClaimBool(claims, "can_read"))
		c.Locals("can_write", jwtutil.ClaimBool(claims, "can_write"))
		c.Locals("can_delete", jwtutil.ClaimBool(claims, "can_delete"))
		c.Locals("can_manage", jwtutil.ClaimBool(claims, "can_manage"))

		return c.Next()
	}
}

func extractTokenFromHeader(header string) (string, error) {
	if header == "" {
		return "", errors.New("empty authorization header")
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("expected format: 'Bearer <token>'")
	}
	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", errors.New("empty token after Bearer prefix")
	}
	return token, nil
}
