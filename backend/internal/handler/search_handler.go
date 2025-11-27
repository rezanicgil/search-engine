// search_handler.go - HTTP handlers for search endpoints
// Handles incoming HTTP requests, validates input, calls services, returns responses
package handler

import (
	"context"
	"search-engine/backend/internal/errors"
	"search-engine/backend/internal/middleware"
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
		// Use custom error type for validation errors
		appErr := errors.NewValidationErrorWithDetails("Invalid request parameters", err.Error())
		middleware.HandleAppError(c, appErr)
		return
	}

	// Perform the search using the service
	// The service handles all business logic and data processing
	// Pass request context for timeout and cancellation support
	response, err := h.searchService.Search(c.Request.Context(), &req)
	if err != nil {
		// Check if it's already an AppError
		if appErr := errors.AsAppError(err); appErr != nil {
			middleware.HandleAppError(c, appErr)
			return
		}

		// Check for context timeout
		if err == context.DeadlineExceeded {
			appErr := errors.NewQueryTimeoutError("search")
			middleware.HandleAppError(c, appErr)
			return
		}

		// Wrap unknown errors
		appErr := errors.NewServiceError("search", err)
		middleware.HandleAppError(c, appErr)
		return
	}

	// Return successful response with search results
	// SearchResponse already has its own structure, so we wrap it in data field
	// for consistency with other endpoints
	middleware.JSONSuccess(c, response)
}
