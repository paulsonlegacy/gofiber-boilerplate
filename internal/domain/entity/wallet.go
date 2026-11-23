package entity

import "time"

// Wallet stores the cached current balance for a user account.
// Balance is always kept in sync with the Transaction ledger — never recomputed from scratch.
// Every credit/debit writes a Transaction AND updates this balance atomically inside a DB transaction.
type Wallet struct {
	ID        string    `gorm:"column:id;primaryKey;size:60"`
	UserID    string    `gorm:"column:user_id;size:60;uniqueIndex;not null"`
	Balance   Money     `gorm:"column:balance;not null;default:0"`
	Currency  string    `gorm:"column:currency;size:3;not null;default:usd"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`

	// Relation
	User *User `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (Wallet) TableName() string {
	return "wallets"
}
