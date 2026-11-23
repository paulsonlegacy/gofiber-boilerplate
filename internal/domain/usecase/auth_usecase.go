// Package usecase contains all business logic for the application.
// Usecases are the heart of the domain layer — they orchestrate repos, infra, and services.
// Handlers call usecases; usecases call repos and services; repos call the DB.
package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"boilerplate/internal/domain/entity"
	"boilerplate/internal/domain/repo"
	"boilerplate/internal/http/apperror"
	jwtutil "boilerplate/internal/infra/jwt"
	"boilerplate/internal/infra/mail"
	"boilerplate/internal/infra/queue"
	"boilerplate/internal/service"

	"gorm.io/gorm"
)

// AuthUsecase handles cross-actor authentication — login, logout, token refresh,
// password reset, and OTP verification for all actor types (user, admin, staff).
type AuthUsecase struct {
	Service   *service.Service
	UserRepo  *repo.UserRepository
	AdminRepo *repo.AdminRepository
}

func NewAuthUsecase(
	svc *service.Service,
	userRepo *repo.UserRepository,
	adminRepo *repo.AdminRepository,
) *AuthUsecase {
	return &AuthUsecase{
		Service:   svc,
		UserRepo:  userRepo,
		AdminRepo: adminRepo,
	}
}

// ── User auth ─────────────────────────────────────────────────────────────────

// LoginUser authenticates a user by email + password and issues token pair.
func (uc *AuthUsecase) LoginUser(ctx context.Context, email, password string) (*entity.User, string, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	password = strings.TrimSpace(password)

	jwtCfg := uc.Service.Config.JWTConfig

	user, err := uc.UserRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, "", "", errors.New("invalid credentials")
	}

	if !jwtutil.Compare(password, user.Password) {
		return nil, "", "", errors.New("invalid credentials")
	}

	if user.Status == entity.ActorStatusSuspended {
		return nil, "", "", errors.New("account has been suspended")
	}
	if user.Status == entity.ActorStatusDeleted {
		return nil, "", "", apperror.New(apperror.ErrCodeAccountDeleted, "Account deleted", "")
	}

	claims := jwtutil.TokenClaims{
		ActorID:     user.ID,
		ActorType:   "user",
		Email:       user.Email,
		DisplayName: user.Name,
		Status:      string(user.Status),
		// All permissions granted by default for account owners.
		// Override per-permission for staff actors.
		CanRead:   true,
		CanWrite:  true,
		CanDelete: true,
		CanManage: true,
	}

	accessToken, tokenRecord, err := jwtutil.MintTokenPair(user.ID, "user", claims, jwtCfg)
	if err != nil {
		return nil, "", "", err
	}

	if err := uc.UserRepo.SaveRefreshToken(ctx, tokenRecord); err != nil {
		return nil, "", "", err
	}

	return user, accessToken, tokenRecord.Token, nil
}

// LogoutUser deletes the refresh token — invalidates the session.
func (uc *AuthUsecase) LogoutUser(ctx context.Context, refreshToken string) error {
	return uc.UserRepo.DeleteRefreshToken(ctx, refreshToken)
}

// RefreshUserToken validates the stored token, re-fetches the user record
// to pick up any status changes, and issues new tokens.
func (uc *AuthUsecase) RefreshUserToken(ctx context.Context, rawToken string) (string, string, error) {
	jwtCfg := uc.Service.Config.JWTConfig

	tokenRecord, err := uc.UserRepo.FindRefreshToken(ctx, rawToken)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}
	if time.Now().After(tokenRecord.ExpiresAt) {
		return "", "", errors.New("refresh token expired")
	}

	user, err := uc.UserRepo.FindByID(ctx, tokenRecord.ActorID)
	if err != nil {
		return "", "", errors.New("user not found")
	}
	if user.Status == entity.ActorStatusSuspended {
		return "", "", errors.New("account has been suspended")
	}

	claims := jwtutil.TokenClaims{
		ActorID:     user.ID,
		ActorType:   "user",
		Email:       user.Email,
		DisplayName: user.Name,
		Status:      string(user.Status),
		CanRead:     true,
		CanWrite:    true,
		CanDelete:   true,
		CanManage:   true,
	}

	newAccess, newRecord, err := jwtutil.MintTokenPair(user.ID, "user", claims, jwtCfg)
	if err != nil {
		return "", "", err
	}

	if err := uc.UserRepo.DeleteRefreshToken(ctx, rawToken); err != nil {
		return "", "", err
	}
	if err := uc.UserRepo.SaveRefreshToken(ctx, newRecord); err != nil {
		return "", "", err
	}

	return newAccess, newRecord.Token, nil
}

// ── Admin auth ────────────────────────────────────────────────────────────────

// LoginAdmin authenticates an admin and issues a token pair.
func (uc *AuthUsecase) LoginAdmin(ctx context.Context, email, password string) (*entity.Admin, string, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	password = strings.TrimSpace(password)

	jwtCfg := uc.Service.Config.JWTConfig

	admin, err := uc.AdminRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, "", "", errors.New("invalid credentials")
	}
	if !jwtutil.Compare(password, admin.Password) {
		return nil, "", "", errors.New("invalid credentials")
	}
	if !admin.IsSuperAdmin && admin.Status == entity.ActorStatusSuspended {
		return nil, "", "", errors.New("account has been deactivated")
	}

	claims := jwtutil.TokenClaims{
		ActorID:      admin.ID,
		ActorType:    "admin",
		Email:        admin.Email,
		DisplayName:  admin.FirstName + " " + admin.LastName,
		IsSuperAdmin: admin.IsSuperAdmin,
		Status:       string(admin.Status),
		CanRead:      true,
		CanWrite:     true,
		CanDelete:    true,
		CanManage:    true,
	}

	accessToken, tokenRecord, err := jwtutil.MintTokenPair(admin.ID, "admin", claims, jwtCfg)
	if err != nil {
		return nil, "", "", err
	}
	if err := uc.AdminRepo.SaveRefreshToken(ctx, tokenRecord); err != nil {
		return nil, "", "", err
	}

	return admin, accessToken, tokenRecord.Token, nil
}

// LogoutAdmin deletes the admin's refresh token.
func (uc *AuthUsecase) LogoutAdmin(ctx context.Context, refreshToken string) error {
	return uc.AdminRepo.DeleteRefreshToken(ctx, refreshToken)
}

// ── Password reset ────────────────────────────────────────────────────────────

// ForgotPassword generates a reset token and queues a password reset email.
func (uc *AuthUsecase) ForgotPassword(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := uc.UserRepo.FindByEmail(ctx, email)
	if err != nil {
		// Return nil — never reveal whether an email exists
		return nil
	}

	token, err := generateResetToken()
	if err != nil {
		return err
	}
	expiresAt := time.Now().Add(1 * time.Hour)

	if err := uc.UserRepo.UpdateFields(ctx, user.ID, map[string]any{
		"reset_token":            token,
		"reset_token_expires_at": expiresAt,
		"updated_at":             time.Now(),
	}); err != nil {
		return err
	}

	resetURL := uc.Service.Config.SiteConfig.BaseURL + "/reset-password?token=" + token

	uc.Service.Queue.Push(queue.NewMailJob(&mail.Email{
		To:       []string{user.Email},
		Subject:  "Reset Your Password",
		Template: "reset_password.html",
		Data: map[string]any{
			"Name":     user.Name,
			"ResetURL": resetURL,
		},
	}))

	return nil
}

// ResetPassword validates the reset token and updates the user's password.
func (uc *AuthUsecase) ResetPassword(ctx context.Context, token, newPassword string) error {
	var user entity.User
	err := uc.Service.DB.WithContext(ctx).
		Where("reset_token = ? AND reset_token_expires_at > ?", token, time.Now()).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperror.NewValidationErr("Invalid or expired reset link")
	}
	if err != nil {
		return err
	}

	return uc.UserRepo.UpdateFields(ctx, user.ID, map[string]any{
		"password":               jwtutil.HashPassword(newPassword, 32),
		"reset_token":            nil,
		"reset_token_expires_at": nil,
		"updated_at":             time.Now(),
	})
}

// ── Internal helpers ──────────────────────────────────────────────────────────

func maskEmail(email string) string {
	at := strings.Index(email, "@")
	if at <= 1 {
		return email
	}
	return string(email[0]) + "***" + email[at:]
}

func generateResetToken() (string, error) {
	b := make([]byte, 32)
	if _, err := jwtutil.CryptoRandRead(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}
