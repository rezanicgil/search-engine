// provider_handler.go - HTTP handlers for provider endpoints
// Handles incoming HTTP requests for provider-related operations

package handler

import (
	"net/http"
	"search-engine/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

// ProviderHandler handles provider-related HTTP requests
type ProviderHandler struct {
	providerRepo *repository.ProviderRepository
}

// NewProviderHandler creates a new ProviderHandler instance
func NewProviderHandler(providerRepo *repository.ProviderRepository) *ProviderHandler {
	return &ProviderHandler{
		providerRepo: providerRepo,
	}
}

// GetProviders handles GET /api/v1/providers requests
// Returns a list of all providers
//
// @Summary     Get providers list
// @Description Get a list of all content providers
// @Tags        providers
// @Accept      json
// @Produce     json
// @Success     200  {array}   model.Provider
// @Failure     500  {object} map[string]string "Internal server error"
// @Router      /providers [get]
func (h *ProviderHandler) GetProviders(c *gin.Context) {
	providers, err := h.providerRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch providers",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": providers,
	})
}
