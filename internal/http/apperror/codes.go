// Package apperror provides structured error types and standard error codes
// for the application layer. Usecases return AppErrors; handlers switch on the
// Code field to map to the correct HTTP status.
package apperror

const (
	// General
	ErrCodeUnknown        = "UNKNOWN_ERROR"
	ErrCodeInternalServer = "INTERNAL_SERVER_ERROR"
	ErrCodeBadRequest     = "BAD_REQUEST"
	ErrCodeTimeout        = "TIMEOUT_ERROR"
	ErrCodeRateLimited    = "RATE_LIMITED"

	// Network / 3rd party
	ErrCodeNetworkError       = "NETWORK_ERROR"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	ErrCodeExternalAPIError   = "EXTERNAL_API_ERROR"

	// Validation / input
	ErrCodeValidation    = "VALIDATION_ERROR"
	ErrCodeMissingField  = "MISSING_FIELD"
	ErrCodeInvalidFormat = "INVALID_FORMAT"
	ErrCodeAlreadyExists = "ALREADY_EXISTS"

	// Authentication / authorization
	ErrCodeAuthFailed         = "AUTH_FAILED"
	ErrCodeAccountDeleted     = "ACCOUNT_DELETED"
	ErrCodeUserConflict       = "USER_CONFLICT"
	ErrCodeUserBanned         = "USER_BANNED"
	ErrCodeUnauthorized       = "UNAUTHORIZED"
	ErrCodeForbidden          = "FORBIDDEN"
	ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"
	ErrCodeTokenExpired       = "TOKEN_EXPIRED"
	ErrCodeTokenInvalid       = "TOKEN_INVALID"

	// Database
	ErrCodeDatabase       = "DATABASE_ERROR"
	ErrCodeRecordNotFound = "RECORD_NOT_FOUND"
	ErrCodeDuplicateKey   = "DUPLICATE_KEY"

	// File uploads
	ErrCodeInvalidFile         = "INVALID_FILE"
	ErrCodeFileTooLarge        = "FILE_TOO_LARGE"
	ErrCodeFileTypeUnsupported = "FILE_TYPE_UNSUPPORTED"
	ErrCodeUploadFailed        = "UPLOAD_FAILED"

	// Payment / wallet
	ErrCodeInsufficientFunds = "INSUFFICIENT_FUNDS"
	ErrCodePaymentFailed     = "PAYMENT_FAILED"
)
