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

// WalletHandler handles wallet balance reads and transaction history.
type WalletHandler struct {
	WalletUC  *usecase.WalletUsecase
	Service   *service.Service
	Validator *validator.Validate
}

func NewWalletHandler(walletUC *usecase.WalletUsecase, svc *service.Service, v *validator.Validate) *WalletHandler {
	return &WalletHandler{WalletUC: walletUC, Service: svc, Validator: v}
}

// GetWallet handles GET /api/wallet
func (h *WalletHandler) GetWallet(c *fiber.Ctx) error {
	wallet, err := h.WalletUC.GetWallet(c.Context(), actorID(c))
	if err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) {
			return response.MapError(c, appErr)
		}
		return response.InternalServerError(c, apperror.NewServerErr(err.Error()))
	}
	return response.Success(c, "Wallet retrieved", map[string]any{
		"id":       wallet.ID,
		"balance":  wallet.Balance.ToMajor(),
		"currency": wallet.Currency,
	})
}

// ListTransactions handles GET /api/wallet/transactions
// For user actors, scope is automatically restricted to their own transactions.
func (h *WalletHandler) ListTransactions(c *fiber.Ctx) error {
	page, limit := paginationParams(c)
	f := filters.TransactionFilters{
		UserID:    actorID(c), // scoped to current user
		Direction: c.Query("direction"),
		Type:      c.Query("type"),
		Status:    c.Query("status"),
		Page:      page,
		Limit:     limit,
	}

	txs, total, err := h.WalletUC.ListTransactions(c.Context(), f)
	if err != nil {
		return response.InternalServerError(c, apperror.NewServerErr(err.Error()))
	}
	return response.Success(c, "Transactions retrieved", response.ToTransactionList(txs, total))
}

// AdminListTransactions handles GET /api/admin/transactions — no scope restriction.
func (h *WalletHandler) AdminListTransactions(c *fiber.Ctx) error {
	page, limit := paginationParams(c)
	f := filters.TransactionFilters{
		UserID:    c.Query("userId"), // admin can optionally filter by user
		Direction: c.Query("direction"),
		Type:      c.Query("type"),
		Status:    c.Query("status"),
		Page:      page,
		Limit:     limit,
	}

	txs, total, err := h.WalletUC.ListTransactions(c.Context(), f)
	if err != nil {
		return response.InternalServerError(c, apperror.NewServerErr(err.Error()))
	}
	return response.Success(c, "Transactions retrieved", response.ToTransactionList(txs, total))
}
