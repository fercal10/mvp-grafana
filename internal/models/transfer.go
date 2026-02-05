package models

import (
	"time"

	"gorm.io/gorm"
)

type Transfer struct {
	ID                uint           `gorm:"primarykey" json:"id"`
	FromAccountID     uint           `gorm:"not null" json:"from_account_id"`
	ToAccountID       uint           `gorm:"not null" json:"to_account_id"`
	Amount            float64        `gorm:"not null" json:"amount"`
	Description       string         `json:"description"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
	
	FromAccount       Account        `gorm:"foreignKey:FromAccountID" json:"from_account,omitempty"`
	ToAccount         Account        `gorm:"foreignKey:ToAccountID" json:"to_account,omitempty"`
}

type CreateTransferRequest struct {
	FromAccountNumber string  `json:"from_account_number" binding:"required"`
	ToAccountNumber   string  `json:"to_account_number" binding:"required"`
	Amount            float64 `json:"amount" binding:"required"`
	Description       string  `json:"description"`
}
