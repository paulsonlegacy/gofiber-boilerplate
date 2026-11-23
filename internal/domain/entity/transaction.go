package entity

import (
	"fmt"
	"time"
)

// Transaction is the unified financial ledger entry.
// Every credit and debit — regardless of source — creates one row here.
// Wallet.Balance is the cached sum; this table is the authoritative history.
type Transaction struct {
	ID     string  `gorm:"column:id;primaryKey;size:60"`
	UserID string  `gorm:"column:user_id;size:60;not null;index"`

	// Polymorphic reference — link to the domain object this transaction relates to.
	// Only one FK is set at a time; infer entity type from which FK is non-nil.
	// Add more FKs here as your domain grows (e.g. SubscriptionID, OrderID).
	RefID   *string `gorm:"column:ref_id;size:60;null;index"`    // generic reference ID
	RefType *string `gorm:"column:ref_type;size:60;null;index"`  // e.g. "order", "subscription"

	Direction TransactionDirection `gorm:"column:direction;size:10;not null"`
	Type      TransactionType      `gorm:"column:type;size:30;not null;index"`
	Method    TransactionMethod    `gorm:"column:method;size:20;not null"`

	Amount   Money  `gorm:"column:amount;not null"`
	Currency string `gorm:"column:currency;size:3;not null;default:usd"`

	// External payment gateway reference — nil when Method = wallet (internal).
	GatewayRef *string `gorm:"column:gateway_ref;size:255;null;index"`

	Notes  *string           `gorm:"column:notes;size:512;null"`
	Status TransactionStatus `gorm:"column:status;size:20;not null;default:completed;index"`

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;index"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`

	// Relation
	User *User `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

func (Transaction) TableName() string {
	return "transactions"
}


// ─── Transaction enums ────────────────────────────────────────────────────────

type TransactionDirection string
type TransactionType      string
type TransactionMethod    string
type TransactionStatus    string

const (
	// Direction
	TransactionDirectionCredit TransactionDirection = "credit"
	TransactionDirectionDebit  TransactionDirection = "debit"

	// Type — extend these to match your domain's money flows
	TransactionTypeTopup       TransactionType = "topup"        // user funded their wallet
	TransactionTypeWithdrawal  TransactionType = "withdrawal"   // user withdrew funds
	TransactionTypeRefund      TransactionType = "refund"       // money returned
	TransactionTypeAdminCredit TransactionType = "admin_credit" // admin manually credited
	TransactionTypeAdminDebit  TransactionType = "admin_debit"  // admin manually debited
	TransactionTypePayment     TransactionType = "payment"      // user paid for something

	// Method — how the money moved
	TransactionMethodWallet      TransactionMethod = "wallet"
	TransactionMethodStripe      TransactionMethod = "stripe"
	TransactionMethodPaystack    TransactionMethod = "paystack"
	TransactionMethodFlutterwave TransactionMethod = "flutterwave"
	TransactionMethodManual      TransactionMethod = "manual"

	// Status
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusReversed  TransactionStatus = "reversed"
)

func (d TransactionDirection) Validate() error {
	switch d {
	case TransactionDirectionCredit, TransactionDirectionDebit:
		return nil
	default:
		return fmt.Errorf("invalid transaction direction: %q", d)
	}
}

func (t TransactionType) Validate() error {
	switch t {
	case TransactionTypeTopup, TransactionTypeWithdrawal, TransactionTypeRefund,
		TransactionTypeAdminCredit, TransactionTypeAdminDebit, TransactionTypePayment:
		return nil
	default:
		return fmt.Errorf("invalid transaction type: %q", t)
	}
}
