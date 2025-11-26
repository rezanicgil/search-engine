// stats_handler.go - HTTP handlers for statistics endpoints
// Handles incoming HTTP requests for system statistics

package handler

import (
	"net/http"
	"search-engine/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

// StatsHandler handles statistics-related HTTP requests
type StatsHandler struct {
	contentRepo  *repository.ContentRepository
	providerRepo *repository.ProviderRepository
}

// NewStatsHandler creates a new StatsHandler instance
func NewStatsHandler(contentRepo *repository.ContentRepository, providerRepo *repository.ProviderRepository) *StatsHandler {
	return &StatsHandler{
		contentRepo:  contentRepo,
		providerRepo: providerRepo,
	}
}

// GetStats handles GET /api/v1/stats requests
// Returns system statistics including content counts, provider info, etc.
//
// @Summary     Get system statistics
// @Description Get statistics about the search engine including content counts, provider information, and type distribution
// @Tags        stats
// @Accept      json
// @Produce     json
// @Success     200  {object} map[string]interface{}
// @Failure     500  {object} map[string]string "Internal server error"
// @Router      /stats [get]
func (h *StatsHandler) GetStats(c *gin.Context) {
	stats, err := h.contentRepo.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch statistics",
			"details": err.Error(),
		})
		return
	}

	// Get provider count
	providers, err := h.providerRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch provider statistics",
			"details": err.Error(),
		})
		return
	}

	stats["providers"] = gin.H{
		"total": len(providers),
		"list":  providers,
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}
