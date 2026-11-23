package usecase

import (
	"context"
	"strings"
	"time"

	"boilerplate/internal/domain/entity"
	"boilerplate/internal/domain/filters"
	"boilerplate/internal/domain/repo"
	"boilerplate/internal/http/apperror"
	jwtutil "boilerplate/internal/infra/jwt"
	"boilerplate/internal/service"

	"github.com/oklog/ulid/v2"
)

// AdminUsecase handles all business logic for admin actors.
// SuperAdmin-only operations are enforced at the middleware level — this layer
// trusts that the caller has already been verified as SuperAdmin.
type AdminUsecase struct {
	Service   *service.Service
	AdminRepo *repo.AdminRepository
}

func NewAdminUsecase(svc *service.Service, adminRepo *repo.AdminRepository) *AdminUsecase {
	return &AdminUsecase{Service: svc, AdminRepo: adminRepo}
}

// ── Admin CRUD ────────────────────────────────────────────────────────────────

// CreateAdmin creates a new admin account. Only callable by a SuperAdmin.
func (uc *AdminUsecase) CreateAdmin(
	ctx context.Context,
	firstName string,
	lastName string,
	email string,
	password string,
	isSuperAdmin bool,
) (*entity.Admin, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	exists, err := uc.AdminRepo.ExistsByField(ctx, "email", email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperror.New(apperror.ErrCodeUserConflict, "An admin with this email already exists", "")
	}

	admin := &entity.Admin{
		ID:           ulid.Make().String(),
		FirstName:    firstName,
		LastName:     lastName,
		Email:        email,
		Password:     jwtutil.HashPassword(password, 32),
		Status:       entity.ActorStatusActive,
		IsSuperAdmin: isSuperAdmin,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := uc.AdminRepo.Create(ctx, admin); err != nil {
		return nil, err
	}

	return admin, nil
}

// GetAdmin returns a single admin by ID.
func (uc *AdminUsecase) GetAdmin(ctx context.Context, adminID string) (*entity.Admin, error) {
	return uc.AdminRepo.FindByID(ctx, adminID)
}

// ListAdmins returns paginated admin accounts.
func (uc *AdminUsecase) ListAdmins(ctx context.Context, f filters.AdminFilters) ([]*entity.Admin, int64, error) {
	return uc.AdminRepo.FindAll(ctx, f)
}

// UpdateAdmin performs a partial update on an admin record.
func (uc *AdminUsecase) UpdateAdmin(ctx context.Context, adminID string, updates map[string]any) error {
	_, err := uc.AdminRepo.FindByID(ctx, adminID)
	if err != nil {
		return apperror.NewNotFoundErr("Admin not found")
	}
	updates["updated_at"] = time.Now()
	return uc.AdminRepo.UpdateFields(ctx, adminID, updates)
}

// DeleteAdmin hard-deletes an admin. Only callable by a SuperAdmin.
// Guards against deleting the last SuperAdmin.
func (uc *AdminUsecase) DeleteAdmin(ctx context.Context, adminID string) error {
	admin, err := uc.AdminRepo.FindByID(ctx, adminID)
	if err != nil {
		return apperror.NewNotFoundErr("Admin not found")
	}
	if admin.IsSuperAdmin {
		count, err := uc.AdminRepo.CountSuperAdmins(ctx)
		if err != nil {
			return err
		}
		if count <= 1 {
			return apperror.NewValidationErr("Cannot delete the last superadmin")
		}
	}
	return uc.AdminRepo.Delete(ctx, adminID)
}

// UpdatePassword verifies the current password then replaces it.
func (uc *AdminUsecase) UpdatePassword(ctx context.Context, adminID, current, newPass string) error {
	admin, err := uc.AdminRepo.FindByID(ctx, adminID)
	if err != nil {
		return apperror.NewNotFoundErr("Admin not found")
	}
	if !jwtutil.Compare(current, admin.Password) {
		return apperror.New(apperror.ErrCodeInvalidCredentials, "Incorrect current password", "")
	}
	return uc.AdminRepo.UpdateFields(ctx, adminID, map[string]any{
		"password":   jwtutil.HashPassword(newPass, 32),
		"updated_at": time.Now(),
	})
}
