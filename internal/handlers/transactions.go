package handlers

import (
	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	ServiceName string
}

func NewTransactionHandler(serviceName string) *TransactionHandler {
	if serviceName == "" {
		serviceName = "api"
	}
	return &TransactionHandler{ServiceName: serviceName}
}

// Health check endpoint
func (h *TransactionHandler) HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "healthy",
		"service": h.ServiceName,
	})
}
