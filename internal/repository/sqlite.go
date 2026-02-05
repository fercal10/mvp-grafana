package repository

import (
	"context"
	"fmt"

	"github.com/tribal/bank-api/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(dbPath string) (*Repository, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto migrate the schema
	if err := db.AutoMigrate(
		&models.Account{},
		&models.Transfer{},
		&models.Transaction{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &Repository{db: db}, nil
}

func (r *Repository) DB() *gorm.DB {
	return r.db
}

// Account operations
func (r *Repository) CreateAccount(ctx context.Context, account *models.Account) error {
	return r.db.WithContext(ctx).Create(account).Error
}

func (r *Repository) GetAccountByID(ctx context.Context, id uint) (*models.Account, error) {
	var account models.Account
	if err := r.db.WithContext(ctx).First(&account, id).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *Repository) GetAccountByNumber(ctx context.Context, accountNumber string) (*models.Account, error) {
	var account models.Account
	if err := r.db.WithContext(ctx).Where("account_number = ?", accountNumber).First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *Repository) ListAccounts(ctx context.Context) ([]models.Account, error) {
	var accounts []models.Account
	if err := r.db.WithContext(ctx).Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *Repository) UpdateAccount(ctx context.Context, account *models.Account) error {
	return r.db.WithContext(ctx).Save(account).Error
}

// Transfer operations
func (r *Repository) CreateTransfer(ctx context.Context, transfer *models.Transfer) error {
	return r.db.WithContext(ctx).Create(transfer).Error
}

func (r *Repository) GetTransferByID(ctx context.Context, id uint) (*models.Transfer, error) {
	var transfer models.Transfer
	if err := r.db.WithContext(ctx).Preload("FromAccount").Preload("ToAccount").First(&transfer, id).Error; err != nil {
		return nil, err
	}
	return &transfer, nil
}

// Transaction operations
func (r *Repository) CreateTransaction(ctx context.Context, transaction *models.Transaction) error {
	return r.db.WithContext(ctx).Create(transaction).Error
}

func (r *Repository) ListTransactionsByAccount(ctx context.Context, accountID uint) ([]models.Transaction, error) {
	var transactions []models.Transaction
	if err := r.db.WithContext(ctx).Where("account_id = ?", accountID).Order("created_at DESC").Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

// Transaction helper for atomic operations
func (r *Repository) WithTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}
