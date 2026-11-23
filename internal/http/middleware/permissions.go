package middleware

import (
	"boilerplate/internal/http/apperror"
	"boilerplate/internal/http/response"

	"github.com/gofiber/fiber/v2"
)

// ActorTypeRequired restricts a route to the specified actor types.
// Pass one or more of: "user", "admin", "staff".
func ActorTypeRequired(allowedTypes ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		actorType, ok := c.Locals("actor_type").(string)
		if !ok || actorType == "" {
			return response.Unauthorized(c, apperror.NewUnauthorizedErr("Actor type could not be identified"))
		}
		for _, allowed := range allowedTypes {
			if actorType == allowed {
				return c.Next()
			}
		}
		return response.Forbidden(c, apperror.NewForbiddenErr("This action is not allowed for your account type"))
	}
}

// RequirePermission gates a route behind a single boolean permission flag.
// Permission keys: "can_read" | "can_write" | "can_delete" | "can_manage"
func RequirePermission(permissionKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		allowed, _ := c.Locals(permissionKey).(bool)
		if !allowed {
			return response.Forbidden(c, apperror.NewForbiddenErr("You do not have permission to perform this action"))
		}
		return c.Next()
	}
}

// AdminOnly restricts a route to admin actors.
func AdminOnly() fiber.Handler {
	return ActorTypeRequired("admin")
}

// UserOnly restricts a route to user actors (account owners).
func UserOnly() fiber.Handler {
	return ActorTypeRequired("user")
}

// UserOrAdmin allows both user owners and admins.
func UserOrAdmin() fiber.Handler {
	return ActorTypeRequired("user", "admin")
}

// SuperAdminOnly restricts a route to admin actors where IsSuperAdmin is true.
func SuperAdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		actorType, _ := c.Locals("actor_type").(string)
		if actorType != "admin" {
			return response.Forbidden(c, apperror.NewForbiddenErr("Admin access required"))
		}
		isSuperAdmin, _ := c.Locals("is_superadmin").(bool)
		if !isSuperAdmin {
			return response.Forbidden(c, apperror.NewForbiddenErr("SuperAdmin access required"))
		}
		return c.Next()
	}
}
