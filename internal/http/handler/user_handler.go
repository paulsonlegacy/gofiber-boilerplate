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

// UserHandler handles requests for user-facing endpoints.
type UserHandler struct {
	UserUC    *usecase.UserUsecase
	Service   *service.Service
	Validator *validator.Validate
}

func NewUserHandler(userUC *usecase.UserUsecase, svc *service.Service, v *validator.Validate) *UserHandler {
	return &UserHandler{UserUC: userUC, Service: svc, Validator: v}
}

// GetProfile handles GET /api/user/profile
func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	user, err := h.UserUC.GetProfile(c.Context(), actorID(c))
	if err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) {
			return response.MapError(c, appErr)
		}
		return response.InternalServerError(c, apperror.NewServerErr(err.Error()))
	}
	return response.Success(c, "Profile retrieved", response.ToUserResponse(user))
}

// UpdateProfile handles PUT /api/user/profile
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	var req struct {
		Name  string `json:"name"  validate:"omitempty,min=2,max=120"`
		Phone string `json:"phone" validate:"omitempty,min=7,max=20"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, apperror.New(apperror.ErrCodeBadRequest, "Invalid request body", err.Error()))
	}
	if err := h.Validator.Struct(req); err != nil {
		return response.BadRequest(c, apperror.New(apperror.ErrCodeValidation, formatValidationError(err), ""))
	}

	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}

	if err := h.UserUC.UpdateProfile(c.Context(), actorID(c), updates); err != nil {
		return response.InternalServerError(c, apperror.NewServerErr(err.Error()))
	}
	return response.Success(c, "Profile updated", nil)
}
