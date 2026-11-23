package apperror

// AppError is the structured error type returned by usecases.
// Handlers use errors.As to detect AppError and switch on Code for HTTP mapping.
type AppError struct {
	Code    string `json:"code"`    // machine-readable constant, e.g. VALIDATION_ERROR
	Message string `json:"message"` // user-facing description
	Details string `json:"details"` // developer-facing context (logged, not sent to client in prod)
}

func (e *AppError) Error() string {
	return e.Details
}

// New constructs an AppError from all three fields.
func New(code, message, details string) *AppError {
	return &AppError{Code: code, Message: message, Details: details}
}

// NewValidationErr is for business rule violations and invalid state transitions.
func NewValidationErr(message string) *AppError {
	return New(ErrCodeValidation, message, "")
}

// NewNotFoundErr is for missing records.
func NewNotFoundErr(details string) *AppError {
	return New(ErrCodeRecordNotFound, "We couldn't find what you're looking for", details)
}

// NewNotFoundErrWithMessage is for when you know what specifically is missing.
func NewNotFoundErrWithMessage(message, details string) *AppError {
	return New(ErrCodeRecordNotFound, message, details)
}

// NewAlreadyExistsErr is for duplicate record conflicts.
func NewAlreadyExistsErr(message string) *AppError {
	return New(ErrCodeAlreadyExists, message, "")
}

// NewDatabaseErr wraps a raw database error. The raw string goes in Details, not Message.
func NewDatabaseErr(details string) *AppError {
	return New(ErrCodeDatabase, "A database error occurred. Please try again.", details)
}

// NewUnauthorizedErr is for missing or invalid authentication.
func NewUnauthorizedErr(details string) *AppError {
	return New(ErrCodeUnauthorized, "Access denied", details)
}

// NewForbiddenErr is for authenticated but unauthorized actions.
func NewForbiddenErr(details string) *AppError {
	return New(ErrCodeForbidden, "You don't have permission to do this", details)
}

// NewInsufficientFundsErr is for wallet balance failures.
func NewInsufficientFundsErr(message string) *AppError {
	return New(ErrCodeInsufficientFunds, message, "")
}

// NewServerErr is a rigid fallback — use when you have nothing specific to say.
func NewServerErr(details string) *AppError {
	return New(ErrCodeInternalServer, "Something went wrong. Please try again.", details)
}

// NewServerErrWithMessage is for when you know what went wrong and can tell the user.
func NewServerErrWithMessage(message, details string) *AppError {
	return New(ErrCodeInternalServer, message, details)
}
