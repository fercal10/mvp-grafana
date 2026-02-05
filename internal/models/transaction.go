package models

import (
	"time"

	"gorm.io/gorm"
)

type TransactionType string

const (
	TransactionTypeDeposit    TransactionType = "deposit"
	TransactionTypeWithdrawal TransactionType = "withdrawal"
	TransactionTypeTransfer   TransactionType = "transfer"
)

type Transaction struct {
	ID          uint            `gorm:"primarykey" json:"id"`
	AccountID   uint            `gorm:"not null;index" json:"account_id"`
	Type        TransactionType `gorm:"not null" json:"type"`
	Amount      float64         `gorm:"not null" json:"amount"`
	Reference   string          `json:"reference"`
	Description string          `json:"description"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"index" json:"-"`
	
	Account     Account         `gorm:"foreignKey:AccountID" json:"account,omitempty"`
}
