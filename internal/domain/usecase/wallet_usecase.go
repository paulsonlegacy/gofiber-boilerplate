package usecase

import (
	"context"
	"time"

	"boilerplate/internal/domain/entity"
	"boilerplate/internal/domain/filters"
	"boilerplate/internal/domain/repo"
	"boilerplate/internal/http/apperror"
	"boilerplate/internal/service"

	"github.com/oklog/ulid/v2"
)

// WalletUsecase handles wallet reads and the transaction ledger.
// All balance mutations go through service.WalletService for atomic DB transactions.
type WalletUsecase struct {
	Service    *service.Service
	WalletRepo *repo.WalletRepository
	TxRepo     *repo.TransactionRepository
	UserRepo   *repo.UserRepository
}

func NewWalletUsecase(
	svc *service.Service,
	walletRepo *repo.WalletRepository,
	txRepo *repo.TransactionRepository,
	userRepo *repo.UserRepository,
) *WalletUsecase {
	return &WalletUsecase{
		Service:    svc,
		WalletRepo: walletRepo,
		TxRepo:     txRepo,
		UserRepo:   userRepo,
	}
}

// GetWallet returns the wallet for the given user.
func (uc *WalletUsecase) GetWallet(ctx context.Context, userID string) (*entity.Wallet, error) {
	wallet, err := uc.WalletRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.NewNotFoundErr("Wallet not found")
	}
	return wallet, nil
}

// ListTransactions returns paginated transactions for a user.
// For admin callers, leave f.UserID empty to return all transactions.
func (uc *WalletUsecase) ListTransactions(ctx context.Context, f filters.TransactionFilters) ([]*entity.Transaction, int64, error) {
	return uc.TxRepo.FindAll(ctx, f)
}

// AdminCreditWallet credits a user's wallet manually and records the transaction.
func (uc *WalletUsecase) AdminCreditWallet(
	ctx context.Context,
	userID string,
	amount entity.Money,
	notes string,
) error {
	wallet, err := uc.WalletRepo.FindByUserID(ctx, userID)
	if err != nil {
		return apperror.NewNotFoundErr("Wallet not found for this user")
	}

	tx := &entity.Transaction{
		ID:        ulid.Make().String(),
		UserID:    userID,
		Direction: entity.TransactionDirectionCredit,
		Type:      entity.TransactionTypeAdminCredit,
		Method:    entity.TransactionMethodManual,
		Amount:    amount,
		Currency:  wallet.Currency,
		Status:    entity.TransactionStatusCompleted,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if notes != "" {
		tx.Notes = &notes
	}

	if err := uc.TxRepo.CreateWithWalletUpdate(ctx, tx, userID, entity.TransactionDirectionCredit, amount); err != nil {
		return apperror.NewDatabaseErr(err.Error())
	}

	return nil
}

// AdminDebitWallet debits a user's wallet manually and records the transaction.
func (uc *WalletUsecase) AdminDebitWallet(
	ctx context.Context,
	userID string,
	amount entity.Money,
	notes string,
) error {
	wallet, err := uc.WalletRepo.FindByUserID(ctx, userID)
	if err != nil {
		return apperror.NewNotFoundErr("Wallet not found for this user")
	}

	if wallet.Balance < amount {
		return apperror.NewInsufficientFundsErr("Insufficient wallet balance")
	}

	tx := &entity.Transaction{
		ID:        ulid.Make().String(),
		UserID:    userID,
		Direction: entity.TransactionDirectionDebit,
		Type:      entity.TransactionTypeAdminDebit,
		Method:    entity.TransactionMethodManual,
		Amount:    amount,
		Currency:  wallet.Currency,
		Status:    entity.TransactionStatusCompleted,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if notes != "" {
		tx.Notes = &notes
	}

	if err := uc.TxRepo.CreateWithWalletUpdate(ctx, tx, userID, entity.TransactionDirectionDebit, amount); err != nil {
		return apperror.NewDatabaseErr(err.Error())
	}

	return nil
}
