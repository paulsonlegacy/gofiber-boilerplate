package service

import (
	"errors"

	"boilerplate/internal/domain/entity"

	"gorm.io/gorm"
)

// WalletService provides low-level atomic wallet operations.
// These are called from usecases that need to mutate balances — never from handlers.
// All mutations happen inside DB transactions to guarantee consistency.
type WalletService struct {
	DB *gorm.DB
}

func NewWalletService(db *gorm.DB) *WalletService {
	return &WalletService{DB: db}
}

// CreditWallet atomically adds amount to the wallet balance.
// Does NOT record a transaction — the caller must do that in the same DB transaction.
//
// tx: pass an existing *gorm.DB transaction, or nil to use the service DB directly.
func (ws *WalletService) CreditWallet(tx *gorm.DB, userID string, amount entity.Money) error {
	if amount <= 0 {
		return errors.New("credit amount must be positive")
	}
	db := ws.DB
	if tx != nil {
		db = tx
	}
	result := db.Model(&entity.Wallet{}).
		Where("user_id = ?", userID).
		Update("balance", gorm.Expr("balance + ?", amount))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("wallet not found")
	}
	return nil
}

// DebitWallet atomically removes amount from the wallet balance with a balance check.
// Returns ErrInsufficientFunds if the balance would go negative.
// Does NOT record a transaction — the caller must do that in the same DB transaction.
func (ws *WalletService) DebitWallet(tx *gorm.DB, userID string, amount entity.Money) error {
	if amount <= 0 {
		return errors.New("debit amount must be positive")
	}
	db := ws.DB
	if tx != nil {
		db = tx
	}

	// Atomic check-and-update: only update if balance >= amount
	result := db.Model(&entity.Wallet{}).
		Where("user_id = ? AND balance >= ?", userID, amount).
		Update("balance", gorm.Expr("balance - ?", amount))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		// Could be wallet not found OR insufficient funds — treat as insufficient
		return errors.New("insufficient wallet balance or wallet not found")
	}
	return nil
}
