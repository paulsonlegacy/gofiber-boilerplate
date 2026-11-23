package repo

import (
	"context"
	"errors"

	"boilerplate/internal/domain/entity"

	"gorm.io/gorm"
)

// WalletRepository provides database operations for Wallet records.
type WalletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) Create(ctx context.Context, wallet *entity.Wallet) error {
	return r.db.WithContext(ctx).Create(wallet).Error
}

func (r *WalletRepository) FindByUserID(ctx context.Context, userID string) (*entity.Wallet, error) {
	var wallet entity.Wallet
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&wallet).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &wallet, err
}

func (r *WalletRepository) UpdateFields(ctx context.Context, id string, fields map[string]any) error {
	return r.db.WithContext(ctx).Model(&entity.Wallet{}).Where("id = ?", id).Updates(fields).Error
}
