package logger

import (
	"boilerplate/internal/http/apperror"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// LogErrorResponse logs a structured error response with request context.
// Call from response helpers to get a consistent audit trail.
func LogErrorResponse(c *fiber.Ctx, label string, appErr *apperror.AppError) {
	zap.L().Warn(label,
		zap.String("method", c.Method()),
		zap.String("path", c.Path()),
		zap.String("ip", c.IP()),
		zap.String("error_code", appErr.Code),
		zap.String("error_message", appErr.Message),
		zap.String("error_details", appErr.Details),
	)
}
