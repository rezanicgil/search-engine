// search_handler.go - HTTP handlers for search endpoints
// Handles incoming HTTP requests, validates input, calls services, returns responses
package handler

import (
	"log"
	"net/http"
	"search-engine/backend/internal/model"
	"search-engine/backend/internal/service"

	"github.com/gin-gonic/gin"
)

// SearchHandler handles search-related HTTP requests
// This struct holds dependencies needed for search operations
type SearchHandler struct {
	searchService *service.SearchService
}

// NewSearchHandler creates a new SearchHandler instance
// This allows dependency injection of the search service
func NewSearchHandler(searchService *service.SearchService) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
	}
}

// Search handles GET /api/v1/search requests
// This endpoint performs content search with filtering, sorting, and pagination
//
// @Summary     Search content
// @Description Search for content with filtering, sorting, and pagination. Results are ranked by relevance score.
// @Tags        search
// @Accept      json
// @Produce     json
// @Param       query        query    string   false  "Search keyword (optional - if empty, returns all content)"
// @Param       type         query    string   false  "Filter by content type: video or article"
// @Param       provider_id  query    int      false  "Filter by provider ID"
// @Param       start_date   query    string   false  "Filter results published on/after this date (YYYY-MM-DD)"
// @Param       end_date     query    string   false  "Filter results published on/before this date (YYYY-MM-DD)"
// @Param       page         query    int      false  "Page number (default: 1)"
// @Param       per_page     query    int      false  "Items per page (default: 10, max: 100)"
// @Param       sort_by      query    string   false  "Sort field: score, published_at, or title (default: score)"
// @Param       sort_order   query    string   false  "Sort order: asc or desc (default: desc)"
// @Success     200          {object} model.SearchResponse
// @Failure     400          {object} map[string]string "Invalid request parameters"
// @Failure     500          {object} map[string]string "Internal server error"
// @Router      /search [get]
func (h *SearchHandler) Search(c *gin.Context) {
	// Bind query parameters to SearchRequest
	// Gin automatically parses query string parameters
	var req model.SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		// Return 400 Bad Request if required parameters are missing or invalid
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	// Perform the search using the service
	// The service handles all business logic and data processing
	response, err := h.searchService.Search(&req)
	if err != nil {
		// Log error for monitoring and debugging
		// In production, you might want to use structured logging
		log.Printf("Search error: %v", err)

		// Return 500 Internal Server Error if search fails
		// Don't expose internal error details in production
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to perform search",
			"message": "An error occurred while processing your search request. Please try again later.",
		})
		return
	}

	// Return successful response with search results
	// Status 200 OK indicates successful search
	c.JSON(http.StatusOK, response)
}
