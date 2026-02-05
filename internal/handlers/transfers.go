package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tribal/bank-api/internal/models"
	"github.com/tribal/bank-api/internal/service"
)

type TransferHandler struct {
	transferService *service.TransferService
}

func NewTransferHandler(transferService *service.TransferService) *TransferHandler {
	return &TransferHandler{
		transferService: transferService,
	}
}

// CreateTransfer godoc
// @Summary Create a new transfer
// @Description Transfer money between accounts
// @Tags transfers
// @Accept json
// @Produce json
// @Param transfer body models.CreateTransferRequest true "Transfer data"
// @Success 201 {object} models.Transfer
// @Router /api/transfers [post]
func (h *TransferHandler) CreateTransfer(c *gin.Context) {
	var req models.CreateTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transfer, err := h.transferService.CreateTransfer(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, transfer)
}

// GetTransfer godoc
// @Summary Get transfer by ID
// @Description Get a single transfer by its ID
// @Tags transfers
// @Accept json
// @Produce json
// @Param id path int true "Transfer ID"
// @Success 200 {object} models.Transfer
// @Router /api/transfers/{id} [get]
func (h *TransferHandler) GetTransfer(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transfer id"})
		return
	}

	transfer, err := h.transferService.GetTransfer(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transfer not found"})
		return
	}

	c.JSON(http.StatusOK, transfer)
}
