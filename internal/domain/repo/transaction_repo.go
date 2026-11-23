package repo

import (
	"context"

	"boilerplate/internal/domain/entity"
	"boilerplate/internal/domain/filters"

	"gorm.io/gorm"
)

// TransactionRepository provides database operations for Transaction records.
type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(ctx context.Context, tx *entity.Transaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

func (r *TransactionRepository) FindByID(ctx context.Context, id string) (*entity.Transaction, error) {
	var tx entity.Transaction
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&tx).Error
	return &tx, err
}

// FindAll returns a paginated list of transactions filtered by the provided filters.
// Handlers scope the UserID field before passing in — admins leave it empty (no scope).
func (r *TransactionRepository) FindAll(ctx context.Context, f filters.TransactionFilters) ([]*entity.Transaction, int64, error) {
	f.ApplyDefaults()

	query := r.db.WithContext(ctx).Model(&entity.Transaction{})

	if f.UserID != "" {
		query = query.Where("user_id = ?", f.UserID)
	}
	if f.Direction != "" {
		query = query.Where("direction = ?", f.Direction)
	}
	if f.Type != "" {
		query = query.Where("type = ?", f.Type)
	}
	if f.Method != "" {
		query = query.Where("method = ?", f.Method)
	}
	if f.Status != "" {
		query = query.Where("status = ?", f.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var txs []*entity.Transaction
	err := query.
		Order(f.SortBy + " " + f.SortDir).
		Limit(f.Limit).
		Offset(f.Offset()).
		Find(&txs).Error

	return txs, total, err
}

// CreateWithWalletUpdate atomically creates a transaction record AND updates the wallet balance.
// Use this for all credit/debit operations — never call them separately.
func (r *TransactionRepository) CreateWithWalletUpdate(
	ctx context.Context,
	tx *entity.Transaction,
	userID string,
	direction entity.TransactionDirection,
	amount entity.Money,
) error {
	return r.db.WithContext(ctx).Transaction(func(db *gorm.DB) error {
		if err := db.Create(tx).Error; err != nil {
			return err
		}

		expr := gorm.Expr("balance + ?", amount)
		if direction == entity.TransactionDirectionDebit {
			expr = gorm.Expr("balance - ?", amount)
		}

		return db.Model(&entity.Wallet{}).
			Where("user_id = ?", userID).
			Update("balance", expr).Error
	})
}
