package wallet

import (
	"fmt"
	"net/http"

	"wallet-go/internal/operation"
	"wallet-go/internal/shared/errors"
	"wallet-go/internal/shared/kafka"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service          *Service
	operationService *operation.Service
	producer         *kafka.Producer
	topicDeposit     string
	topicWithdraw    string
	topicTransfer    string
}

func NewHandler(service *Service, operationService *operation.Service, producer *kafka.Producer, topicDeposit, topicWithdraw, topicTransfer string) *Handler {
	return &Handler{
		service:          service,
		operationService: operationService,
		producer:         producer,
		topicDeposit:     topicDeposit,
		topicWithdraw:    topicWithdraw,
		topicTransfer:    topicTransfer,
	}
}

// Create godoc
// @Summary Create wallet
// @Description Create a new wallet for a customer
// @Tags Wallet
// @Accept json
// @Produce json
// @Param request body object{customer_id=string} true "Wallet creation request"
// @Success 201 {object} object{id=string,customer_id=string,current_amount_in_cents=int,active=bool,created_at=string}
// @Failure 400 {object} object{error=string,message=string}
// @Failure 500 {object} object{error=string,message=string}
// @Router /wallet [post]
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

// GetByID godoc
// @Summary Get wallet by ID
// @Description Get wallet details by wallet ID
// @Tags Wallet
// @Accept json
// @Produce json
// @Param id path string true "Wallet ID"
// @Success 200 {object} object{id=string,customer_id=string,current_amount_in_cents=int,active=bool,operations=array}
// @Failure 400 {object} object{error=string,message=string}
// @Failure 404 {object} object{error=string,message=string}
// @Failure 500 {object} object{error=string,message=string}
// @Router /wallet/{id} [get]
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

	// Carregar operações da carteira
	operations, err := h.operationService.GetByWalletID(c.Request.Context(), walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.InternalServerError("Failed to load wallet operations"))
		return
	}

	// Adicionar operações à carteira
	wallet.Operations = operations

	response := h.mapToResponse(wallet)
	c.JSON(http.StatusOK, response)
}

// List godoc
// @Summary List wallets
// @Description List all wallets
// @Tags Wallet
// @Accept json
// @Produce json
// @Success 200 {array} object{id=string,customer_id=string,current_amount_in_cents=int,active=bool}
// @Failure 500 {object} object{error=string,message=string}
// @Router /wallet [get]
func (h *Handler) List(c *gin.Context) {
	wallets, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.InternalServerError("Failed to list wallets"))
		return
	}

	responses := make([]WalletResponse, len(wallets))
	for i, wallet := range wallets {
		// Carregar operações para cada carteira
		operations, err := h.operationService.GetByWalletID(c.Request.Context(), wallet.WalletID)
		if err != nil {
			fmt.Printf("DEBUG: Erro ao carregar operações para wallet %s: %v\n", wallet.WalletID, err)
			// Não quebrar por causa das operações - apenas não incluir
		} else {
			wallet.Operations = operations
		}

		responses[i] = *h.mapToResponse(wallet)
	}

	c.JSON(http.StatusOK, responses)
}

// Patch godoc
// @Summary Update wallet
// @Description Update wallet status (active/blocked)
// @Tags Wallet
// @Accept json
// @Produce json
// @Param id path string true "Wallet ID"
// @Param request body object{active=bool,blocked=bool} true "Wallet patch request"
// @Success 200 {object} object{id=string,customer_id=string,current_amount_in_cents=int,active=bool,blocked=bool}
// @Failure 400 {object} object{error=string,message=string}
// @Failure 404 {object} object{error=string,message=string}
// @Failure 500 {object} object{error=string,message=string}
// @Router /wallet/{id} [patch]
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

// Deposit godoc
// @Summary Deposit to wallet
// @Description Deposit money to wallet (async operation)
// @Tags Wallet
// @Accept json
// @Produce json
// @Param id path string true "Wallet ID"
// @Param request body object{amount_in_cents=int} true "Deposit request"
// @Success 202 {object} object{message=string}
// @Failure 400 {object} object{error=string,message=string}
// @Failure 500 {object} object{error=string,message=string}
// @Router /wallet/{id}/deposit [post]
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

// Withdraw godoc
// @Summary Withdraw from wallet
// @Description Withdraw money from wallet (async operation)
// @Tags Wallet
// @Accept json
// @Produce json
// @Param id path string true "Wallet ID"
// @Param request body object{amount_in_cents=int} true "Withdraw request"
// @Success 202 {object} object{message=string}
// @Failure 400 {object} object{error=string,message=string}
// @Failure 500 {object} object{error=string,message=string}
// @Router /wallet/{id}/withdraw [post]
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

// Transfer godoc
// @Summary Transfer between wallets
// @Description Transfer money between wallets (async operation)
// @Tags Wallet
// @Accept json
// @Produce json
// @Param id path string true "Source Wallet ID"
// @Param request body object{amount_in_cents=int,wallet_destination_id=string} true "Transfer request"
// @Success 202 {object} object{message=string}
// @Failure 400 {object} object{error=string,message=string}
// @Failure 500 {object} object{error=string,message=string}
// @Router /wallet/{id}/transfer [post]
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

// mapToResponse mapper Wallet to WalletResponse
func (h *Handler) mapToResponse(wallet *Wallet) *WalletResponse {
	return &WalletResponse{
		Id:                   wallet.WalletID,
		CustomerID:           wallet.CustomerID,
		CurrentAmountInCents: wallet.CurrentAmountInCents,
		Operations:           wallet.Operations,
		Active:               wallet.Active,
		Blocked:              wallet.Blocked,
		CreatedAt:            wallet.CreatedAt,
		UpdatedAt:            wallet.UpdatedAt,
		BlockedAt:            wallet.BlockedAt,
		UnblockedAt:          wallet.UnblockedAt,
	}
}
