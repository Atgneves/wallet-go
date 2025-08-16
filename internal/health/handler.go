package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Health(c *gin.Context) {
	health := h.service.GetHealth()

	statusCode := http.StatusOK
	if health.Status == HealthStatusDown {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, health)
}

func (h *Handler) HealthDetails(c *gin.Context) {
	health := h.service.GetHealthWithDetails()

	statusCode := http.StatusOK
	if health.Status == HealthStatusDown {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, health)
}
