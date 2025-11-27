// content_handler.go - HTTP handlers for content endpoints
// Handles incoming HTTP requests for content-related operations

package handler

import (
	"context"
	"search-engine/backend/internal/errors"
	"search-engine/backend/internal/middleware"
	"search-engine/backend/internal/repository"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ContentHandler handles content-related HTTP requests
type ContentHandler struct {
	contentRepo        *repository.ContentRepository
	simpleQueryTimeout time.Duration
}

// NewContentHandler creates a new ContentHandler instance
// simpleQueryTimeout is the timeout for simple queries like GetByID (default: 5s)
func NewContentHandler(contentRepo *repository.ContentRepository, simpleQueryTimeout time.Duration) *ContentHandler {
	if simpleQueryTimeout <= 0 {
		simpleQueryTimeout = 5 * time.Second
	}
	return &ContentHandler{
		contentRepo:        contentRepo,
		simpleQueryTimeout: simpleQueryTimeout,
	}
}

// GetContentByID handles GET /api/v1/content/:id requests
// Returns detailed information about a specific content item
//
// @Summary     Get content by ID
// @Description Get detailed information about a specific content item by its ID
// @Tags        content
// @Accept      json
// @Produce     json
// @Param       id   path     int  true  "Content ID"
// @Success     200  {object} model.Content
// @Failure     400  {object} map[string]string "Invalid content ID"
// @Failure     404  {object} map[string]string "Content not found"
// @Failure     500  {object} map[string]string "Internal server error"
// @Router      /content/{id} [get]
func (h *ContentHandler) GetContentByID(c *gin.Context) {
	// Parse content ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		appErr := errors.NewInvalidIDError("content")
		middleware.HandleAppError(c, appErr)
		return
	}

	// Apply timeout for simple query
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.simpleQueryTimeout)
	defer cancel()

	// Get content from repository with context for timeout and cancellation
	content, err := h.contentRepo.GetByID(ctx, id)
	if err != nil {
		// Check for timeout
		if ctx.Err() == context.DeadlineExceeded {
			appErr := errors.NewRequestTimeoutErrorWithDuration(h.simpleQueryTimeout.String())
			middleware.HandleAppError(c, appErr)
			return
		}

		// Check for not found first (before checking if it's AppError)
		// This allows us to add details (like ID) to the error
		if err == repository.ErrContentNotFound || err == errors.ErrContentNotFound {
			appErr := errors.NewContentNotFoundErrorWithID(id)
			middleware.HandleAppError(c, appErr)
			return
		}

		// Check if it's already an AppError
		if appErr := errors.AsAppError(err); appErr != nil {
			// If it's a not found error without details, add ID
			if appErr.Code == errors.ErrorCodeContentNotFound && appErr.Details == "" {
				appErr = errors.NewContentNotFoundErrorWithID(id)
			}
			middleware.HandleAppError(c, appErr)
			return
		}

		// Wrap unknown errors
		appErr := errors.NewDatabaseError("get content by id", err)
		middleware.HandleAppError(c, appErr)
		return
	}

	// Load tags for the content (use same timeout)
	tags, err := h.contentRepo.GetTagsByContentID(ctx, id)
	if err != nil {
		// Log error but don't fail the request
		// Tags are optional metadata
	} else {
		content.Tags = tags
	}

	middleware.JSONSuccess(c, content)
}
