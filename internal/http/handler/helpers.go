package handler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// formatValidationError converts go-playground validator errors into a single
// human-readable sentence. Returns the first field error found.
func formatValidationError(err error) string {
	if ve, ok := err.(validator.ValidationErrors); ok && len(ve) > 0 {
		e := ve[0]
		return fmt.Sprintf("Field '%s' failed validation: %s", strings.ToLower(e.Field()), e.Tag())
	}
	return "Validation failed"
}

// paginationParams extracts page and limit from query params with safe defaults.
func paginationParams(c *fiber.Ctx) (page, limit int) {
	page, _ = strconv.Atoi(c.Query("page", "1"))
	limit, _ = strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return
}

// actorID safely reads the actor ID from Locals.
func actorID(c *fiber.Ctx) string {
	id, _ := c.Locals("actor_id").(string)
	return id
}

// actorType safely reads the actor type from Locals.
func actorType(c *fiber.Ctx) string {
	t, _ := c.Locals("actor_type").(string)
	return t
}

// isAdmin returns true if the current actor is an admin.
func isAdmin(c *fiber.Ctx) bool {
	return actorType(c) == "admin"
}

// isSuperAdmin returns true if the current actor is a superadmin.
func isSuperAdmin(c *fiber.Ctx) bool {
	v, _ := c.Locals("is_superadmin").(bool)
	return v
}
