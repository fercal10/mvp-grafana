package models

import (
	"time"

	"gorm.io/gorm"
)

type Account struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	AccountNumber string         `gorm:"uniqueIndex;not null" json:"account_number"`
	Balance       float64        `gorm:"not null;default:0" json:"balance"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type CreateAccountRequest struct {
	AccountNumber string  `json:"account_number" binding:"required"`
	InitialBalance float64 `json:"initial_balance"`
}
