package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tribal/bank-api/internal/models"
	"github.com/tribal/bank-api/internal/service"
)

type AccountHandler struct {
	accountService *service.AccountService
}

func NewAccountHandler(accountService *service.AccountService) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
	}
}

// ListAccounts godoc
// @Summary List all accounts
// @Description Get a list of all bank accounts
// @Tags accounts
// @Accept json
// @Produce json
// @Success 200 {array} models.Account
// @Router /api/accounts [get]
func (h *AccountHandler) ListAccounts(c *gin.Context) {
	accounts, err := h.accountService.ListAccounts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, accounts)
}

// GetAccount godoc
// @Summary Get account by ID
// @Description Get a single account by its ID
// @Tags accounts
// @Accept json
// @Produce json
// @Param id path int true "Account ID"
// @Success 200 {object} models.Account
// @Router /api/accounts/{id} [get]
func (h *AccountHandler) GetAccount(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}

	account, err := h.accountService.GetAccount(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}

	c.JSON(http.StatusOK, account)
}

// CreateAccount godoc
// @Summary Create a new account
// @Description Create a new bank account
// @Tags accounts
// @Accept json
// @Produce json
// @Param account body models.CreateAccountRequest true "Account data"
// @Success 201 {object} models.Account
// @Router /api/accounts [post]
func (h *AccountHandler) CreateAccount(c *gin.Context) {
	var req models.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := h.accountService.CreateAccount(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, account)
}

// GetAccountTransactions godoc
// @Summary Get account transactions
// @Description Get all transactions for a specific account
// @Tags accounts
// @Accept json
// @Produce json
// @Param id path int true "Account ID"
// @Success 200 {array} models.Transaction
// @Router /api/accounts/{id}/transactions [get]
func (h *AccountHandler) GetAccountTransactions(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}

	transactions, err := h.accountService.GetAccountTransactions(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transactions)
}
