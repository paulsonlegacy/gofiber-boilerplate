package response

import (
	"boilerplate/internal/domain/entity"
	"time"
)

// TransactionResponse is the public representation of a Transaction.
type TransactionResponse struct {
	ID         string    `json:"id"`
	Direction  string    `json:"direction"`
	Type       string    `json:"type"`
	Method     string    `json:"method"`
	Amount     float64   `json:"amount"` // major unit (e.g. dollars)
	Currency   string    `json:"currency"`
	Status     string    `json:"status"`
	GatewayRef *string   `json:"gatewayRef,omitempty"`
	Notes      *string   `json:"notes,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

func ToTransactionResponse(t *entity.Transaction) *TransactionResponse {
	return &TransactionResponse{
		ID:         t.ID,
		Direction:  string(t.Direction),
		Type:       string(t.Type),
		Method:     string(t.Method),
		Amount:     t.Amount.ToMajor(),
		Currency:   t.Currency,
		Status:     string(t.Status),
		GatewayRef: t.GatewayRef,
		Notes:      t.Notes,
		CreatedAt:  t.CreatedAt,
	}
}

func ToTransactionList(txs []*entity.Transaction, total int64) map[string]any {
	items := make([]*TransactionResponse, 0, len(txs))
	for _, t := range txs {
		items = append(items, ToTransactionResponse(t))
	}
	return map[string]any{
		"items": items,
		"total": total,
	}
}
