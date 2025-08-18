package operation

import (
	"net/http"
	"time"

	"wallet-go/internal/shared/errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetByID(c *gin.Context) {
	idParam := c.Param("operationId")
	operationID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.BadRequest("Invalid operation ID"))
		return
	}

	operation, err := h.service.GetByID(c.Request.Context(), operationID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.Code, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, errors.InternalServerError("Failed to get operation"))
		return
	}

	response := h.mapToResponse(operation)
	c.JSON(http.StatusOK, response)
}

func (h *Handler) List(c *gin.Context) {
	var request OperationFilterRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		c.JSON(http.StatusBadRequest, errors.BadRequest("Invalid query parameters"))
		return
	}

	operations, err := h.service.List(c.Request.Context(), request)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.Code, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, errors.InternalServerError("Failed to list operations"))
		return
	}

	responses := make([]OperationResponse, len(operations))
	for i, operation := range operations {
		responses[i] = *h.mapToResponse(operation)
	}

	c.JSON(http.StatusOK, responses)
}

// GetDailySummary godoc
// @Summary Get wallet balance
// @Description Get wallet balance from today's operations
// @Tags Wallet
// @Accept json
// @Produce json
// @Success 200 {object} object{date=string,total_deposits=int,total_withdraws=int,total_transfers=int}
// @Failure 500 {object} object{error=string,message=string}
// @Router /wallet/daily-summary [get]
func (h *Handler) GetDailySummary(c *gin.Context) {
	walletIDParam := c.Query("walletId")
	if walletIDParam == "" {
		c.JSON(http.StatusBadRequest, errors.BadRequest("walletId is required"))
		return
	}

	walletID, err := uuid.Parse(walletIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.BadRequest("Invalid wallet ID"))
		return
	}

	dateParam := c.Query("date")
	if dateParam == "" {
		c.JSON(http.StatusBadRequest, errors.BadRequest("date is required"))
		return
	}

	date, err := time.Parse("2006-01-02", dateParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.BadRequest("Invalid date format. Use YYYY-MM-DD"))
		return
	}

	summary, err := h.service.GetDailySummary(c.Request.Context(), walletID, date)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.Code, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, errors.InternalServerError("Failed to get daily summary"))
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetDailySummaryDetails godoc
// @Summary Get daily wallet balance details
// @Description Get summary of daily operations for wallets
// @Tags Wallet
// @Accept json
// @Produce json
// @Success 200 {object} object{date=string,total_deposits=int,total_withdraws=int,total_transfers=int}
// @Failure 500 {object} object{error=string,message=string}
// @Router /wallet/daily-summary-details [get]
func (h *Handler) GetDailySummaryDetails(c *gin.Context) {
	walletIDParam := c.Query("walletId")
	if walletIDParam == "" {
		c.JSON(http.StatusBadRequest, errors.BadRequest("walletId is required"))
		return
	}

	walletID, err := uuid.Parse(walletIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.BadRequest("Invalid wallet ID"))
		return
	}

	dateParam := c.Query("date")
	if dateParam == "" {
		c.JSON(http.StatusBadRequest, errors.BadRequest("date is required"))
		return
	}

	date, err := time.Parse("2006-01-02", dateParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.BadRequest("Invalid date format. Use YYYY-MM-DD"))
		return
	}

	summary, err := h.service.GetDailySummaryDetails(c.Request.Context(), walletID, date)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.Code, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, errors.InternalServerError("Failed to get daily summary details"))
		return
	}

	c.JSON(http.StatusOK, summary)
}

func (h *Handler) mapToResponse(operation *Operation) *OperationResponse {
	return &OperationResponse{
		ID:                     operation.OperationID,
		WalletID:               operation.WalletID,
		Type:                   operation.Type,
		Status:                 operation.Status,
		AmountInCents:          operation.AmountInCents,
		WalletTransactionID:    operation.WalletTransactionID,
		OperationTransactionID: operation.OperationTransactionID,
		Reason:                 operation.Reason,
		CreatedAt:              operation.CreatedAt,
		UpdatedAt:              operation.UpdatedAt,
	}
}
