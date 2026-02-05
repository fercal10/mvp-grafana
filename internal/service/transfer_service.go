package service

import (
	"context"
	"fmt"

	"github.com/tribal/bank-api/internal/models"
	"github.com/tribal/bank-api/internal/repository"
	"github.com/tribal/bank-api/pkg/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

type TransferService struct {
	repo *repository.Repository
}

func NewTransferService(repo *repository.Repository) *TransferService {
	return &TransferService{repo: repo}
}

func (s *TransferService) CreateTransfer(ctx context.Context, req models.CreateTransferRequest) (*models.Transfer, error) {
	ctx, span := tracer.Start(ctx, "TransferService.CreateTransfer")
	defer span.End()

	span.SetAttributes(
		attribute.String("transfer.from", req.FromAccountNumber),
		attribute.String("transfer.to", req.ToAccountNumber),
		attribute.Float64("transfer.amount", req.Amount),
	)

	var transfer *models.Transfer
	var fromAccount, toAccount *models.Account

	// Execute transfer in a transaction
	err := s.repo.WithTransaction(ctx, func(tx *gorm.DB) error {
		// Get source account
		var err error
		fromAccount, err = s.repo.GetAccountByNumber(ctx, req.FromAccountNumber)
		if err != nil {
			return fmt.Errorf("source account not found: %w", err)
		}

		// Get destination account
		toAccount, err = s.repo.GetAccountByNumber(ctx, req.ToAccountNumber)
		if err != nil {
			return fmt.Errorf("destination account not found: %w", err)
		}

		// Check if accounts are different
		if fromAccount.ID == toAccount.ID {
			return fmt.Errorf("cannot transfer to the same account")
		}

		// Check sufficient balance (minimal validation as requested)
		if fromAccount.Balance < req.Amount {
			return fmt.Errorf("insufficient balance")
		}

		// Update balances
		fromAccount.Balance -= req.Amount
		toAccount.Balance += req.Amount

		if err := tx.Save(fromAccount).Error; err != nil {
			return fmt.Errorf("failed to update source account: %w", err)
		}

		if err := tx.Save(toAccount).Error; err != nil {
			return fmt.Errorf("failed to update destination account: %w", err)
		}

		// Create transfer record
		transfer = &models.Transfer{
			FromAccountID: fromAccount.ID,
			ToAccountID:   toAccount.ID,
			Amount:        req.Amount,
			Description:   req.Description,
		}

		if err := tx.Create(transfer).Error; err != nil {
			return fmt.Errorf("failed to create transfer record: %w", err)
		}

		// Create transactions for both accounts
		reference := fmt.Sprintf("TRF-%d", transfer.ID)

		withdrawalTx := &models.Transaction{
			AccountID:   fromAccount.ID,
			Type:        models.TransactionTypeTransfer,
			Amount:      -req.Amount,
			Reference:   reference,
			Description: fmt.Sprintf("Transfer to %s", toAccount.AccountNumber),
		}

		if err := tx.Create(withdrawalTx).Error; err != nil {
			return fmt.Errorf("failed to create withdrawal transaction: %w", err)
		}

		depositTx := &models.Transaction{
			AccountID:   toAccount.ID,
			Type:        models.TransactionTypeTransfer,
			Amount:      req.Amount,
			Reference:   reference,
			Description: fmt.Sprintf("Transfer from %s", fromAccount.AccountNumber),
		}

		if err := tx.Create(depositTx).Error; err != nil {
			return fmt.Errorf("failed to create deposit transaction: %w", err)
		}

		return nil
	})

	if err != nil {
		span.RecordError(err)
		// Record failed transfer
		telemetry.RecordTransfer(req.Amount, false)
		return nil, err
	}

	// Record successful transfer
	telemetry.RecordTransfer(req.Amount, true)
	telemetry.UpdateAccountBalance(fromAccount.AccountNumber, fromAccount.Balance)
	telemetry.UpdateAccountBalance(toAccount.AccountNumber, toAccount.Balance)

	// Load the full transfer with related accounts
	transfer.FromAccount = *fromAccount
	transfer.ToAccount = *toAccount

	return transfer, nil
}

func (s *TransferService) GetTransfer(ctx context.Context, id uint) (*models.Transfer, error) {
	ctx, span := tracer.Start(ctx, "TransferService.GetTransfer")
	defer span.End()

	span.SetAttributes(attribute.Int("transfer.id", int(id)))

	transfer, err := s.repo.GetTransferByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get transfer: %w", err)
	}

	return transfer, nil
}
