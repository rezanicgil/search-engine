// content_handler.go - HTTP handlers for content endpoints
// Handles incoming HTTP requests for content-related operations

package handler

import (
	"log"
	"net/http"
	"search-engine/backend/internal/repository"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ContentHandler handles content-related HTTP requests
type ContentHandler struct {
	contentRepo *repository.ContentRepository
}

// NewContentHandler creates a new ContentHandler instance
func NewContentHandler(contentRepo *repository.ContentRepository) *ContentHandler {
	return &ContentHandler{
		contentRepo: contentRepo,
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid content ID",
			"details": "Content ID must be a valid integer",
		})
		return
	}

	// Get content from repository
	content, err := h.contentRepo.GetByID(id)
	if err != nil {
		if err == repository.ErrContentNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Content not found",
				"message": "No content found with the specified ID",
			})
			return
		}
		// Log error for monitoring
		log.Printf("Error fetching content %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch content",
			"message": "An error occurred while fetching content. Please try again later.",
		})
		return
	}

	// Load tags for the content
	tags, err := h.contentRepo.GetTagsByContentID(id)
	if err != nil {
		// Log error but don't fail the request
		// Tags are optional metadata
	} else {
		content.Tags = tags
	}

	c.JSON(http.StatusOK, gin.H{
		"data": content,
	})
}
