package wallet

import (
	"net/http"

	"wallet-go/internal/shared/errors"
	"wallet-go/internal/shared/kafka"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service       *Service
	producer      *kafka.Producer
	topicDeposit  string
	topicWithdraw string
	topicTransfer string
}

func NewHandler(service *Service, producer *kafka.Producer, topicDeposit, topicWithdraw, topicTransfer string) *Handler {
	return &Handler{
		service:       service,
		producer:      producer,
		topicDeposit:  topicDeposit,
		topicWithdraw: topicWithdraw,
		topicTransfer: topicTransfer,
	}
}

func (h *Handler) Create(c *gin.Context) {
	var request WalletRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, errors.BadRequest("Invalid request body"))
		return
	}

	wallet, err := h.service.Create(c.Request.Context(), request)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.Code, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, errors.InternalServerError("Failed to create wallet"))
		return
	}

	response := h.mapToResponse(wallet)
	c.JSON(http.StatusCreated, response)
}

func (h *Handler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	walletID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.BadRequest("Invalid wallet ID"))
		return
	}

	wallet, err := h.service.GetByID(c.Request.Context(), walletID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.Code, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, errors.InternalServerError("Failed to get wallet"))
		return
	}

	response := h.mapToResponse(wallet)
	c.JSON(http.StatusOK, response)
}

func (h *Handler) List(c *gin.Context) {
	wallets, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.InternalServerError("Failed to list wallets"))
		return
	}

	responses := make([]WalletResponse, len(wallets))
	for i, wallet := range wallets {
		responses[i] = *h.mapToResponse(wallet)
	}

	c.JSON(http.StatusOK, responses)
}

func (h *Handler) Patch(c *gin.Context) {
	idParam := c.Param("id")
	walletID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.BadRequest("Invalid wallet ID"))
		return
	}

	var patch WalletPatch
	if err := c.ShouldBindJSON(&patch); err != nil {
		c.JSON(http.StatusBadRequest, errors.BadRequest("Invalid request body"))
		return
	}

	wallet, err := h.service.Patch(c.Request.Context(), walletID, patch)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.Code, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, errors.InternalServerError("Failed to update wallet"))
		return
	}

	response := h.mapToResponse(wallet)
	c.JSON(http.StatusOK, response)
}

func (h *Handler) Deposit(c *gin.Context) {
	idParam := c.Param("id")
	walletID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.BadRequest("Invalid wallet ID"))
		return
	}

	var request WalletTransactionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, errors.BadRequest("Invalid request body"))
		return
	}

	message := WalletKafkaTransactionMessage{
		WalletID:      walletID,
		AmountInCents: request.AmountInCents,
	}

	if err := h.producer.SendMessage(c.Request.Context(), h.topicDeposit, walletID.String(), message); err != nil {
		c.JSON(http.StatusInternalServerError, errors.InternalServerError("Failed to send deposit message"))
		return
	}

	response := map[string]string{
		"message": "Deposit request accepted!",
	}
	c.JSON(http.StatusAccepted, response)
}

func (h *Handler) Withdraw(c *gin.Context) {
	idParam := c.Param("id")
	walletID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.BadRequest("Invalid wallet ID"))
		return
	}

	var request WalletTransactionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, errors.BadRequest("Invalid request body"))
		return
	}

	message := WalletKafkaTransactionMessage{
		WalletID:      walletID,
		AmountInCents: request.AmountInCents,
	}

	if err := h.producer.SendMessage(c.Request.Context(), h.topicWithdraw, walletID.String(), message); err != nil {
		c.JSON(http.StatusInternalServerError, errors.InternalServerError("Failed to send withdraw message"))
		return
	}

	response := map[string]string{
		"message": "Withdrawal request accepted!",
	}
	c.JSON(http.StatusAccepted, response)
}

func (h *Handler) Transfer(c *gin.Context) {
	idParam := c.Param("id")
	walletID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.BadRequest("Invalid wallet ID"))
		return
	}

	var request WalletTransactionTransferRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, errors.BadRequest("Invalid request body"))
		return
	}

	message := WalletKafkaTransactionTransferMessage{
		WalletID:            walletID,
		AmountInCents:       request.AmountInCents,
		WalletDestinationID: request.WalletDestinationID,
	}

	if err := h.producer.SendMessage(c.Request.Context(), h.topicTransfer, walletID.String(), message); err != nil {
		c.JSON(http.StatusInternalServerError, errors.InternalServerError("Failed to send transfer message"))
		return
	}

	response := map[string]string{
		"message": "Transfer request accepted!",
	}
	c.JSON(http.StatusAccepted, response)
}

func (h *Handler) mapToResponse(wallet *Wallet) *WalletResponse {
	return &WalletResponse{
		ID:                   wallet.WalletID,
		CustomerID:           wallet.CustomerID,
		CurrentAmountInCents: wallet.CurrentAmountInCents,
		Active:               wallet.Active,
		Blocked:              wallet.Blocked,
		CreatedAt:            wallet.CreatedAt,
		UpdatedAt:            wallet.UpdatedAt,
		BlockedAt:            wallet.BlockedAt,
		UnblockedAt:          wallet.UnblockedAt,
	}
}
