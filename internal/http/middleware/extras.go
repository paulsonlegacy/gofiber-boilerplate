package middleware

import (
	"boilerplate/internal/http/apperror"
	"boilerplate/internal/http/response"
	"boilerplate/internal/util"

	"github.com/gofiber/fiber/v2"
)

// SanitizeInputs cleans URL params and query strings to help prevent XSS/injection.
// Cleaned values are stored in Locals under "param_<name>" and "query_<name>".
func SanitizeInputs() fiber.Handler {
	return func(c *fiber.Ctx) error {
		for _, param := range c.Route().Params {
			if original := c.Params(param); original != "" {
				c.Locals("param_"+param, util.SanitizeInput(original))
			}
		}
		c.Request().URI().QueryArgs().VisitAll(func(key, value []byte) {
			c.Locals("query_"+string(key), util.SanitizeInput(string(value)))
		})
		return c.Next()
	}
}

// MaintenanceMode blocks all requests with a 503 when enabled is true.
// Toggle via config or an admin endpoint — no restart required.
func MaintenanceMode(enabled bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if enabled {
			return response.ServiceUnavailable(c, apperror.New(
				apperror.ErrCodeServiceUnavailable,
				"The system is undergoing maintenance. Please check back shortly.",
				"maintenance mode is active",
			))
		}
		return c.Next()
	}
}
