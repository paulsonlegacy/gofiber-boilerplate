// Package handler contains Fiber HTTP handlers — the outermost application layer.
// Handlers are thin: parse request → call usecase → map response. No business logic here.
// Error mapping follows: usecase returns *apperror.AppError → handler switches on Code → HTTP response.
package handler

import (
	"errors"

	"boilerplate/internal/domain/usecase"
	"boilerplate/internal/http/apperror"
	"boilerplate/internal/http/response"
	"boilerplate/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// AuthHandler handles authentication for all actor types.
type AuthHandler struct {
	AuthUC    *usecase.AuthUsecase
	UserUC    *usecase.UserUsecase
	Service   *service.Service
	Validator *validator.Validate
}

func NewAuthHandler(authUC *usecase.AuthUsecase, userUC *usecase.UserUsecase, svc *service.Service, v *validator.Validate) *AuthHandler {
	return &AuthHandler{AuthUC: authUC, UserUC: userUC, Service: svc, Validator: v}
}

// Register handles POST /api/auth/register
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req struct {
		Name     string `json:"name"     validate:"required,min=2,max=120"`
		Email    string `json:"email"    validate:"required,email"`
		Phone    string `json:"phone"    validate:"required,min=7,max=20"`
		Password string `json:"password" validate:"required,min=8"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, apperror.New(apperror.ErrCodeBadRequest, "Invalid request body", err.Error()))
	}
	if err := h.Validator.Struct(req); err != nil {
		return response.BadRequest(c, apperror.New(apperror.ErrCodeValidation, formatValidationError(err), ""))
	}

	user, err := h.UserUC.Register(c.Context(), req.Name, req.Email, req.Phone, req.Password)
	if err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) {
			return response.MapError(c, appErr)
		}
		return response.InternalServerError(c, apperror.NewServerErr(err.Error()))
	}

	return response.Created(c, "Account created successfully", response.ToUserResponse(user))
}

// Login handles POST /api/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"    validate:"required"`
		Password string `json:"password" validate:"required"`
		Role     string `json:"role"` // "user" | "admin"
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, apperror.New(apperror.ErrCodeBadRequest, "Invalid request body", err.Error()))
	}
	if err := h.Validator.Struct(req); err != nil {
		return response.BadRequest(c, apperror.New(apperror.ErrCodeValidation, formatValidationError(err), ""))
	}

	switch req.Role {
	case "admin":
		admin, access, refresh, err := h.AuthUC.LoginAdmin(c.Context(), req.Email, req.Password)
		if err != nil {
			var appErr *apperror.AppError
			if errors.As(err, &appErr) {
				return response.MapError(c, appErr)
			}
			return response.Unauthorized(c, apperror.New(apperror.ErrCodeInvalidCredentials, err.Error(), ""))
		}
		return response.Success(c, "Login successful", map[string]any{
			"actor":        response.ToAdminResponse(admin),
			"accessToken":  access,
			"refreshToken": refresh,
		})

	default: // "user"
		user, access, refresh, err := h.AuthUC.LoginUser(c.Context(), req.Email, req.Password)
		if err != nil {
			var appErr *apperror.AppError
			if errors.As(err, &appErr) {
				return response.MapError(c, appErr)
			}
			return response.Unauthorized(c, apperror.New(apperror.ErrCodeInvalidCredentials, err.Error(), ""))
		}
		return response.Success(c, "Login successful", map[string]any{
			"actor":        response.ToUserResponse(user),
			"accessToken":  access,
			"refreshToken": refresh,
		})
	}
}

// Logout handles POST /api/auth/logout (or /api/user/logout)
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refreshToken" validate:"required"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, apperror.New(apperror.ErrCodeBadRequest, "Invalid request body", err.Error()))
	}

	actorType, _ := c.Locals("actor_type").(string)
	switch actorType {
	case "admin":
		_ = h.AuthUC.LogoutAdmin(c.Context(), req.RefreshToken)
	default:
		_ = h.AuthUC.LogoutUser(c.Context(), req.RefreshToken)
	}

	return response.Success(c, "Logged out successfully", nil)
}

// RefreshToken handles POST /api/auth/refresh-token
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refreshToken" validate:"required"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, apperror.New(apperror.ErrCodeBadRequest, "Invalid request body", err.Error()))
	}
	if err := h.Validator.Struct(req); err != nil {
		return response.BadRequest(c, apperror.New(apperror.ErrCodeValidation, formatValidationError(err), ""))
	}

	access, refresh, err := h.AuthUC.RefreshUserToken(c.Context(), req.RefreshToken)
	if err != nil {
		return response.Unauthorized(c, apperror.New(apperror.ErrCodeTokenInvalid, err.Error(), ""))
	}

	return response.Success(c, "Token refreshed", map[string]any{
		"accessToken":  access,
		"refreshToken": refresh,
	})
}

// ForgotPassword handles POST /api/auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email" validate:"required,email"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, apperror.New(apperror.ErrCodeBadRequest, "Invalid request body", err.Error()))
	}
	if err := h.Validator.Struct(req); err != nil {
		return response.BadRequest(c, apperror.New(apperror.ErrCodeValidation, formatValidationError(err), ""))
	}

	_ = h.AuthUC.ForgotPassword(c.Context(), req.Email)
	// Always return success — never confirm whether the email exists
	return response.Success(c, "If an account exists with this email, you'll receive a reset link shortly", nil)
}

// ResetPassword handles POST /api/auth/reset-password
func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req struct {
		Token       string `json:"token"       validate:"required"`
		NewPassword string `json:"newPassword" validate:"required,min=8"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, apperror.New(apperror.ErrCodeBadRequest, "Invalid request body", err.Error()))
	}
	if err := h.Validator.Struct(req); err != nil {
		return response.BadRequest(c, apperror.New(apperror.ErrCodeValidation, formatValidationError(err), ""))
	}

	if err := h.AuthUC.ResetPassword(c.Context(), req.Token, req.NewPassword); err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) {
			return response.MapError(c, appErr)
		}
		return response.BadRequest(c, apperror.NewValidationErr("Invalid or expired reset link"))
	}

	return response.Success(c, "Password reset successfully", nil)
}
