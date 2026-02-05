package handlers

import (
	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	// Transactions are handled through AccountHandler's GetAccountTransactions
	// This file is a placeholder for any future transaction-specific endpoints
}

func NewTransactionHandler() *TransactionHandler {
	return &TransactionHandler{}
}

// Health check endpoint
func (h *TransactionHandler) HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "healthy",
		"service": "bank-api",
	})
}
