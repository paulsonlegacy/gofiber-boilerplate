// Package repo contains GORM-backed repository implementations.
// Repositories are the only layer allowed to touch the database directly.
// Usecases call repos; handlers never touch repos directly.
package repo

import (
	"context"
	"errors"

	"boilerplate/internal/domain/entity"
	"boilerplate/internal/domain/filters"

	"gorm.io/gorm"
)

// UserRepository provides all database operations for the User entity.
type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// ── Write operations ──────────────────────────────────────────────────────────

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) UpdateFields(ctx context.Context, id string, fields map[string]any) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).Updates(fields).Error
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.User{}, "id = ?", id).Error
}

// ── Read operations ───────────────────────────────────────────────────────────

func (r *UserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &user, err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &user, err
}

func (r *UserRepository) FindAll(ctx context.Context, f filters.UserFilters) ([]*entity.User, int64, error) {
	f.ApplyDefaults()

	query := r.db.WithContext(ctx).Model(&entity.User{})

	if f.Status != "" {
		query = query.Where("status = ?", f.Status)
	}
	if f.Search != "" {
		like := "%" + f.Search + "%"
		query = query.Where("name LIKE ? OR email LIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []*entity.User
	err := query.
		Order(f.SortBy + " " + f.SortDir).
		Limit(f.Limit).
		Offset(f.Offset()).
		Find(&users).Error

	return users, total, err
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.User{}).Where("phone = ?", phone).Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) ExistsByField(ctx context.Context, field, value string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.User{}).Where(field+" = ?", value).Count(&count).Error
	return count > 0, err
}

// ── Refresh token operations ──────────────────────────────────────────────────

func (r *UserRepository) SaveRefreshToken(ctx context.Context, token *entity.RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *UserRepository) FindRefreshToken(ctx context.Context, raw string) (*entity.RefreshToken, error) {
	var token entity.RefreshToken
	err := r.db.WithContext(ctx).Where("token = ? AND actor_type = ?", raw, "user").First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &token, err
}

func (r *UserRepository) DeleteRefreshToken(ctx context.Context, raw string) error {
	return r.db.WithContext(ctx).Where("token = ?", raw).Delete(&entity.RefreshToken{}).Error
}

func (r *UserRepository) DeleteExpiredRefreshTokens(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("actor_type = ? AND expires_at < NOW()", "user").
		Delete(&entity.RefreshToken{}).Error
}
