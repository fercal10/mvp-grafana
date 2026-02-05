package service

import (
	"context"
	"fmt"

	"github.com/tribal/bank-api/internal/models"
	"github.com/tribal/bank-api/internal/repository"
	"github.com/tribal/bank-api/pkg/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var tracer = otel.Tracer("bank-api")

type AccountService struct {
	repo *repository.Repository
}

func NewAccountService(repo *repository.Repository) *AccountService {
	return &AccountService{repo: repo}
}

func (s *AccountService) CreateAccount(ctx context.Context, req models.CreateAccountRequest) (*models.Account, error) {
	ctx, span := tracer.Start(ctx, "AccountService.CreateAccount")
	defer span.End()

	span.SetAttributes(attribute.String("account.number", req.AccountNumber))

	account := &models.Account{
		AccountNumber: req.AccountNumber,
		Balance:       req.InitialBalance,
	}

	if err := s.repo.CreateAccount(ctx, account); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	// Record Prometheus metric
	telemetry.RecordAccountCreation()
	telemetry.UpdateAccountBalance(account.AccountNumber, account.Balance)

	// Create initial transaction if there's an initial balance
	if req.InitialBalance > 0 {
		transaction := &models.Transaction{
			AccountID:   account.ID,
			Type:        models.TransactionTypeDeposit,
			Amount:      req.InitialBalance,
			Description: "Initial deposit",
			Reference:   fmt.Sprintf("INIT-%d", account.ID),
		}
		if err := s.repo.CreateTransaction(ctx, transaction); err != nil {
			span.RecordError(err)
			// Log error but don't fail account creation
		}
	}

	return account, nil
}

func (s *AccountService) GetAccount(ctx context.Context, id uint) (*models.Account, error) {
	ctx, span := tracer.Start(ctx, "AccountService.GetAccount")
	defer span.End()

	span.SetAttributes(attribute.Int("account.id", int(id)))

	account, err := s.repo.GetAccountByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return account, nil
}

func (s *AccountService) ListAccounts(ctx context.Context) ([]models.Account, error) {
	ctx, span := tracer.Start(ctx, "AccountService.ListAccounts")
	defer span.End()

	accounts, err := s.repo.ListAccounts(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}

	span.SetAttributes(attribute.Int("accounts.count", len(accounts)))

	return accounts, nil
}

func (s *AccountService) GetAccountTransactions(ctx context.Context, id uint) ([]models.Transaction, error) {
	ctx, span := tracer.Start(ctx, "AccountService.GetAccountTransactions")
	defer span.End()

	span.SetAttributes(attribute.Int("account.id", int(id)))

	// First verify account exists
	if _, err := s.repo.GetAccountByID(ctx, id); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("account not found: %w", err)
	}

	transactions, err := s.repo.ListTransactionsByAccount(ctx, id)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	span.SetAttributes(attribute.Int("transactions.count", len(transactions)))

	return transactions, nil
}
