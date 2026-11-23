// Package response provides standard JSON response helpers for all HTTP handlers.
// All API responses follow the same envelope:
//
//	{ "message": string, "success": bool, "data": any, "error": { "code", "message", "details" } }
//
// Handlers never construct raw JSON maps — they always call one of these helpers.
package response

import (
	"boilerplate/internal/http/apperror"
	"boilerplate/internal/infra/logger"

	"github.com/gofiber/fiber/v2"
)

// ── Success responses ─────────────────────────────────────────────────────────

func Success(c *fiber.Ctx, message string, data any) error {
	if message == "" {
		message = "Success"
	}
	payload := normalizeData(data)
	return c.Status(fiber.StatusOK).JSON(envelope(message, true, payload, emptyError()))
}

func Created(c *fiber.Ctx, message string, data any) error {
	if message == "" {
		message = "Created successfully"
	}
	payload := normalizeData(data)
	return c.Status(fiber.StatusCreated).JSON(envelope(message, true, payload, emptyError()))
}

// ── Error responses ───────────────────────────────────────────────────────────

func BadRequest(c *fiber.Ctx, appErr *apperror.AppError) error {
	appErr = fallback(appErr, "BAD_REQUEST", "Bad request")
	logger.LogErrorResponse(c, "Bad Request", appErr)
	return c.Status(fiber.StatusBadRequest).JSON(errEnvelope(appErr))
}

func Unauthorized(c *fiber.Ctx, appErr *apperror.AppError) error {
	appErr = fallback(appErr, "UNAUTHORIZED", "Unauthorized")
	logger.LogErrorResponse(c, "Unauthorized", appErr)
	return c.Status(fiber.StatusUnauthorized).JSON(errEnvelope(appErr))
}

func Forbidden(c *fiber.Ctx, appErr *apperror.AppError) error {
	appErr = fallback(appErr, "FORBIDDEN", "Forbidden")
	logger.LogErrorResponse(c, "Forbidden", appErr)
	return c.Status(fiber.StatusForbidden).JSON(errEnvelope(appErr))
}

func NotFound(c *fiber.Ctx, appErr *apperror.AppError) error {
	appErr = fallback(appErr, "NOT_FOUND", "Not found")
	return c.Status(fiber.StatusNotFound).JSON(errEnvelope(appErr))
}

func Conflict(c *fiber.Ctx, appErr *apperror.AppError) error {
	appErr = fallback(appErr, "CONFLICT", "Conflict")
	logger.LogErrorResponse(c, "Conflict", appErr)
	return c.Status(fiber.StatusConflict).JSON(errEnvelope(appErr))
}

func Unprocessable(c *fiber.Ctx, appErr *apperror.AppError) error {
	appErr = fallback(appErr, "UNPROCESSABLE", "Unprocessable entity")
	return c.Status(fiber.StatusUnprocessableEntity).JSON(errEnvelope(appErr))
}

func InternalServerError(c *fiber.Ctx, appErr *apperror.AppError) error {
	appErr = fallback(appErr, "INTERNAL_SERVER_ERROR", "An error occurred")
	logger.LogErrorResponse(c, "Internal Server Error", appErr)
	return c.Status(fiber.StatusInternalServerError).JSON(errEnvelope(appErr))
}

func ServiceUnavailable(c *fiber.Ctx, appErr *apperror.AppError) error {
	appErr = fallback(appErr, "SERVICE_UNAVAILABLE", "Service unavailable")
	logger.LogErrorResponse(c, "Service Unavailable", appErr)
	return c.Status(fiber.StatusServiceUnavailable).JSON(errEnvelope(appErr))
}

func PaymentRequired(c *fiber.Ctx, appErr *apperror.AppError) error {
	appErr = fallback(appErr, "INSUFFICIENT_FUNDS", "Insufficient funds")
	return c.Status(fiber.StatusPaymentRequired).JSON(errEnvelope(appErr))
}

// ── AppError → HTTP status mapper ─────────────────────────────────────────────

// MapError maps an AppError code to the correct HTTP response.
// Use this in handlers instead of manual switches for consistent error handling.
func MapError(c *fiber.Ctx, err *apperror.AppError) error {
	switch err.Code {
	case apperror.ErrCodeRecordNotFound:
		return NotFound(c, err)
	case apperror.ErrCodeValidation, apperror.ErrCodeMissingField, apperror.ErrCodeInvalidFormat,
		apperror.ErrCodeInvalidCredentials:
		return BadRequest(c, err)
	case apperror.ErrCodeUnauthorized, apperror.ErrCodeTokenExpired, apperror.ErrCodeTokenInvalid,
		apperror.ErrCodeAccountDeleted:
		return Unauthorized(c, err)
	case apperror.ErrCodeForbidden:
		return Forbidden(c, err)
	case apperror.ErrCodeAlreadyExists, apperror.ErrCodeUserConflict, apperror.ErrCodeDuplicateKey:
		return Conflict(c, err)
	case apperror.ErrCodeInsufficientFunds:
		return PaymentRequired(c, err)
	default:
		return InternalServerError(c, err)
	}
}

// ── Internal helpers ──────────────────────────────────────────────────────────

func envelope(message string, success bool, data, errBlock any) map[string]any {
	return map[string]any{
		"message": message,
		"success": success,
		"data":    data,
		"error":   errBlock,
	}
}

func errEnvelope(appErr *apperror.AppError) map[string]any {
	return envelope(appErr.Message, false, emptyData(), fiber.Map{
		"code":    appErr.Code,
		"message": appErr.Message,
		"details": appErr.Details,
	})
}

func emptyError() map[string]any { return map[string]any{} }
func emptyData() map[string]any  { return map[string]any{} }

func normalizeData(data any) any {
	if data == nil {
		return emptyData()
	}
	return data
}

func fallback(appErr *apperror.AppError, code, msg string) *apperror.AppError {
	if appErr == nil {
		return apperror.New(code, msg, "")
	}
	return appErr
}
