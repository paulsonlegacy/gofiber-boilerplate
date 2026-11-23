package handler

import (
	"errors"

	"boilerplate/internal/domain/filters"
	"boilerplate/internal/domain/usecase"
	"boilerplate/internal/http/apperror"
	"boilerplate/internal/http/response"
	"boilerplate/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// AdminHandler handles admin management and user management endpoints.
type AdminHandler struct {
	AdminUC   *usecase.AdminUsecase
	UserUC    *usecase.UserUsecase
	Service   *service.Service
	Validator *validator.Validate
}

func NewAdminHandler(adminUC *usecase.AdminUsecase, userUC *usecase.UserUsecase, svc *service.Service, v *validator.Validate) *AdminHandler {
	return &AdminHandler{AdminUC: adminUC, UserUC: userUC, Service: svc, Validator: v}
}

// ── User management (admin) ───────────────────────────────────────────────────

// ListUsers handles GET /api/admin/users
func (h *AdminHandler) ListUsers(c *fiber.Ctx) error {
	page, limit := paginationParams(c)
	f := filters.UserFilters{
		Status:  c.Query("status"),
		Search:  c.Query("search"),
		Page:    page,
		Limit:   limit,
		SortBy:  c.Query("sortBy", "created_at"),
		SortDir: c.Query("sortDir", "desc"),
	}

	users, total, err := h.UserUC.ListUsers(c.Context(), f)
	if err != nil {
		return response.InternalServerError(c, apperror.NewServerErr(err.Error()))
	}
	return response.Success(c, "Users retrieved", response.ToUserList(users, total))
}

// GetUser handles GET /api/admin/users/:id
func (h *AdminHandler) GetUser(c *fiber.Ctx) error {
	user, err := h.UserUC.GetUser(c.Context(), c.Params("id"))
	if err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) {
			return response.MapError(c, appErr)
		}
		return response.InternalServerError(c, apperror.NewServerErr(err.Error()))
	}
	return response.Success(c, "User retrieved", response.ToUserResponse(user))
}

// UpdateUser handles PUT /api/admin/users/:id
func (h *AdminHandler) UpdateUser(c *fiber.Ctx) error {
	var req struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, apperror.New(apperror.ErrCodeBadRequest, "Invalid request body", err.Error()))
	}

	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	if err := h.UserUC.UpdateProfile(c.Context(), c.Params("id"), updates); err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) {
			return response.MapError(c, appErr)
		}
		return response.InternalServerError(c, apperror.NewServerErr(err.Error()))
	}
	return response.Success(c, "User updated", nil)
}

// SuspendUser handles POST /api/admin/users/:id/suspend
func (h *AdminHandler) SuspendUser(c *fiber.Ctx) error {
	if err := h.UserUC.SuspendUser(c.Context(), c.Params("id")); err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) {
			return response.MapError(c, appErr)
		}
		return response.InternalServerError(c, apperror.NewServerErr(err.Error()))
	}
	return response.Success(c, "User suspended", nil)
}

// ActivateUser handles POST /api/admin/users/:id/activate
func (h *AdminHandler) ActivateUser(c *fiber.Ctx) error {
	if err := h.UserUC.ActivateUser(c.Context(), c.Params("id")); err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) {
			return response.MapError(c, appErr)
		}
		return response.InternalServerError(c, apperror.NewServerErr(err.Error()))
	}
	return response.Success(c, "User activated", nil)
}

// DeleteUser handles DELETE /api/admin/users/:id
func (h *AdminHandler) DeleteUser(c *fiber.Ctx) error {
	if err := h.UserUC.DeleteUser(c.Context(), c.Params("id")); err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) {
			return response.MapError(c, appErr)
		}
		return response.InternalServerError(c, apperror.NewServerErr(err.Error()))
	}
	return response.Success(c, "User deleted", nil)
}

// ── Admin management (superadmin) ─────────────────────────────────────────────

// ListAdmins handles GET /api/superadmin/admins
func (h *AdminHandler) ListAdmins(c *fiber.Ctx) error {
	page, limit := paginationParams(c)
	f := filters.AdminFilters{
		Status: c.Query("status"),
		Search: c.Query("search"),
		Page:   page,
		Limit:  limit,
	}

	admins, total, err := h.AdminUC.ListAdmins(c.Context(), f)
	if err != nil {
		return response.InternalServerError(c, apperror.NewServerErr(err.Error()))
	}
	return response.Success(c, "Admins retrieved", response.ToAdminList(admins, total))
}

// CreateAdmin handles POST /api/superadmin/admins
func (h *AdminHandler) CreateAdmin(c *fiber.Ctx) error {
	var req struct {
		FirstName    string `json:"firstName"    validate:"required"`
		LastName     string `json:"lastName"     validate:"required"`
		Email        string `json:"email"        validate:"required,email"`
		Password     string `json:"password"     validate:"required,min=8"`
		IsSuperAdmin bool   `json:"isSuperAdmin"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, apperror.New(apperror.ErrCodeBadRequest, "Invalid request body", err.Error()))
	}
	if err := h.Validator.Struct(req); err != nil {
		return response.BadRequest(c, apperror.New(apperror.ErrCodeValidation, formatValidationError(err), ""))
	}

	admin, err := h.AdminUC.CreateAdmin(c.Context(), req.FirstName, req.LastName, req.Email, req.Password, req.IsSuperAdmin)
	if err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) {
			return response.MapError(c, appErr)
		}
		return response.InternalServerError(c, apperror.NewServerErr(err.Error()))
	}
	return response.Created(c, "Admin created", response.ToAdminResponse(admin))
}

// UpdateAdmin handles PUT /api/superadmin/admins/:id
func (h *AdminHandler) UpdateAdmin(c *fiber.Ctx) error {
	var req struct {
		FirstName    string `json:"firstName"`
		LastName     string `json:"lastName"`
		Email        string `json:"email"`
		Status       string `json:"status"`
		IsSuperAdmin *bool  `json:"isSuperAdmin"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, apperror.New(apperror.ErrCodeBadRequest, "Invalid request body", err.Error()))
	}

	updates := map[string]any{}
	if req.FirstName != "" {
		updates["first_name"] = req.FirstName
	}
	if req.LastName != "" {
		updates["last_name"] = req.LastName
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.IsSuperAdmin != nil {
		updates["is_superadmin"] = *req.IsSuperAdmin
	}

	if err := h.AdminUC.UpdateAdmin(c.Context(), c.Params("id"), updates); err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) {
			return response.MapError(c, appErr)
		}
		return response.InternalServerError(c, apperror.NewServerErr(err.Error()))
	}
	return response.Success(c, "Admin updated", nil)
}

// DeleteAdmin handles DELETE /api/superadmin/admins/:id
func (h *AdminHandler) DeleteAdmin(c *fiber.Ctx) error {
	if err := h.AdminUC.DeleteAdmin(c.Context(), c.Params("id")); err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) {
			return response.MapError(c, appErr)
		}
		return response.InternalServerError(c, apperror.NewServerErr(err.Error()))
	}
	return response.Success(c, "Admin deleted", nil)
}
