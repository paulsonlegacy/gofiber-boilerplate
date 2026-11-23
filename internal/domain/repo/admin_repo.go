package repo

import (
	"context"
	"errors"

	"boilerplate/internal/domain/entity"
	"boilerplate/internal/domain/filters"

	"gorm.io/gorm"
)

// AdminRepository provides all database operations for the Admin entity.
type AdminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

// ── Write operations ──────────────────────────────────────────────────────────

func (r *AdminRepository) Create(ctx context.Context, admin *entity.Admin) error {
	return r.db.WithContext(ctx).Create(admin).Error
}

func (r *AdminRepository) UpdateFields(ctx context.Context, id string, fields map[string]any) error {
	return r.db.WithContext(ctx).Model(&entity.Admin{}).Where("id = ?", id).Updates(fields).Error
}

func (r *AdminRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Admin{}, "id = ?", id).Error
}

// ── Read operations ───────────────────────────────────────────────────────────

func (r *AdminRepository) FindByID(ctx context.Context, id string) (*entity.Admin, error) {
	var admin entity.Admin
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&admin).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &admin, err
}

func (r *AdminRepository) FindByEmail(ctx context.Context, email string) (*entity.Admin, error) {
	var admin entity.Admin
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&admin).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &admin, err
}

func (r *AdminRepository) FindAll(ctx context.Context, f filters.AdminFilters) ([]*entity.Admin, int64, error) {
	f.ApplyDefaults()

	query := r.db.WithContext(ctx).Model(&entity.Admin{})

	if f.Status != "" {
		query = query.Where("status = ?", f.Status)
	}
	if f.IsSuperAdmin != nil {
		query = query.Where("is_superadmin = ?", *f.IsSuperAdmin)
	}
	if f.Search != "" {
		like := "%" + f.Search + "%"
		query = query.Where("first_name LIKE ? OR last_name LIKE ? OR email LIKE ?", like, like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var admins []*entity.Admin
	err := query.
		Order("created_at desc").
		Limit(f.Limit).
		Offset(f.Offset()).
		Find(&admins).Error

	return admins, total, err
}

func (r *AdminRepository) ExistsByField(ctx context.Context, field, value string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Admin{}).Where(field+" = ?", value).Count(&count).Error
	return count > 0, err
}

func (r *AdminRepository) CountSuperAdmins(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Admin{}).Where("is_superadmin = ?", true).Count(&count).Error
	return count, err
}

// ── Refresh token operations ──────────────────────────────────────────────────

func (r *AdminRepository) SaveRefreshToken(ctx context.Context, token *entity.RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *AdminRepository) FindRefreshToken(ctx context.Context, raw string) (*entity.RefreshToken, error) {
	var token entity.RefreshToken
	err := r.db.WithContext(ctx).Where("token = ? AND actor_type = ?", raw, "admin").First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &token, err
}

func (r *AdminRepository) DeleteRefreshToken(ctx context.Context, raw string) error {
	return r.db.WithContext(ctx).Where("token = ?", raw).Delete(&entity.RefreshToken{}).Error
}

func (r *AdminRepository) DeleteExpiredRefreshTokens(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("actor_type = ? AND expires_at < NOW()", "admin").
		Delete(&entity.RefreshToken{}).Error
}
