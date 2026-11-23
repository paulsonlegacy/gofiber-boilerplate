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

// UserUsecase handles all business logic for user actors.
type UserUsecase struct {
	Service    *service.Service
	UserRepo   *repo.UserRepository
	WalletRepo *repo.WalletRepository
	TxRepo     *repo.TransactionRepository
}

func NewUserUsecase(
	svc *service.Service,
	userRepo *repo.UserRepository,
	walletRepo *repo.WalletRepository,
	txRepo *repo.TransactionRepository,
) *UserUsecase {
	return &UserUsecase{
		Service:    svc,
		UserRepo:   userRepo,
		WalletRepo: walletRepo,
		TxRepo:     txRepo,
	}
}

// ── Registration ──────────────────────────────────────────────────────────────

// Register creates a new user account and provisions a wallet.
// Returns the created user or an AppError for conflict/validation failures.
func (uc *UserUsecase) Register(
	ctx context.Context,
	name string,
	email string,
	phone string,
	password string,
) (*entity.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	phone = strings.TrimSpace(phone)

	emailExists, err := uc.UserRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, apperror.NewDatabaseErr(err.Error())
	}
	if emailExists {
		return nil, apperror.New(apperror.ErrCodeAlreadyExists, "An account with this email already exists", "")
	}

	phoneExists, err := uc.UserRepo.ExistsByPhone(ctx, phone)
	if err != nil {
		return nil, apperror.NewDatabaseErr(err.Error())
	}
	if phoneExists {
		return nil, apperror.New(apperror.ErrCodeAlreadyExists, "An account with this phone number already exists", "")
	}

	user := &entity.User{
		ID:        ulid.Make().String(),
		Name:      name,
		Email:     email,
		Phone:     phone,
		Password:  jwtutil.HashPassword(password, 32),
		Status:    entity.ActorStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := uc.UserRepo.Create(ctx, user); err != nil {
		return nil, apperror.NewDatabaseErr(err.Error())
	}

	// Provision wallet automatically on registration
	wallet := &entity.Wallet{
		ID:       ulid.Make().String(),
		UserID:   user.ID,
		Balance:  0,
		Currency: "usd", // adapt to your project's currency
	}
	if err := uc.WalletRepo.Create(ctx, wallet); err != nil {
		return nil, apperror.NewDatabaseErr(err.Error())
	}

	return user, nil
}

// ── Profile ───────────────────────────────────────────────────────────────────

// GetProfile returns a user by ID.
func (uc *UserUsecase) GetProfile(ctx context.Context, userID string) (*entity.User, error) {
	user, err := uc.UserRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, apperror.NewNotFoundErr("User not found")
	}
	return user, nil
}

// UpdateProfile applies a partial update to user profile fields.
func (uc *UserUsecase) UpdateProfile(ctx context.Context, userID string, updates map[string]any) error {
	updates["updated_at"] = time.Now()
	return uc.UserRepo.UpdateFields(ctx, userID, updates)
}

// UpdatePassword verifies the current password then replaces it.
func (uc *UserUsecase) UpdatePassword(ctx context.Context, userID, current, newPass string) error {
	user, err := uc.UserRepo.FindByID(ctx, userID)
	if err != nil {
		return apperror.NewNotFoundErr("User not found")
	}
	if !jwtutil.Compare(current, user.Password) {
		return apperror.New(apperror.ErrCodeInvalidCredentials, "Incorrect current password", "")
	}
	return uc.UserRepo.UpdateFields(ctx, userID, map[string]any{
		"password":   jwtutil.HashPassword(newPass, 32),
		"updated_at": time.Now(),
	})
}

// ── Admin-facing user management ──────────────────────────────────────────────

// ListUsers returns paginated users — called by admin handlers.
func (uc *UserUsecase) ListUsers(ctx context.Context, f filters.UserFilters) ([]*entity.User, int64, error) {
	return uc.UserRepo.FindAll(ctx, f)
}

// GetUser returns a single user — called by admin handlers.
func (uc *UserUsecase) GetUser(ctx context.Context, userID string) (*entity.User, error) {
	user, err := uc.UserRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, apperror.NewNotFoundErr("User not found")
	}
	return user, nil
}

// SuspendUser sets user status to suspended.
func (uc *UserUsecase) SuspendUser(ctx context.Context, userID string) error {
	_, err := uc.UserRepo.FindByID(ctx, userID)
	if err != nil {
		return apperror.NewNotFoundErr("User not found")
	}
	return uc.UserRepo.UpdateFields(ctx, userID, map[string]any{
		"status":     entity.ActorStatusSuspended,
		"updated_at": time.Now(),
	})
}

// ActivateUser sets user status to active.
func (uc *UserUsecase) ActivateUser(ctx context.Context, userID string) error {
	_, err := uc.UserRepo.FindByID(ctx, userID)
	if err != nil {
		return apperror.NewNotFoundErr("User not found")
	}
	return uc.UserRepo.UpdateFields(ctx, userID, map[string]any{
		"status":     entity.ActorStatusActive,
		"updated_at": time.Now(),
	})
}

// DeleteUser soft-deletes a user by marking status as deleted.
func (uc *UserUsecase) DeleteUser(ctx context.Context, userID string) error {
	_, err := uc.UserRepo.FindByID(ctx, userID)
	if err != nil {
		return apperror.NewNotFoundErr("User not found")
	}
	return uc.UserRepo.UpdateFields(ctx, userID, map[string]any{
		"status":     entity.ActorStatusDeleted,
		"updated_at": time.Now(),
	})
}
