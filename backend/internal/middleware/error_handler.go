// error_handler.go - Centralized error handling middleware
// Provides consistent error responses across all endpoints
package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"search-engine/backend/internal/errors"

	"github.com/gin-gonic/gin"
)

// ErrorHandlerMiddleware returns a middleware that handles errors consistently
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			handleError(c, err)
			return
		}
	}
}

// handleError processes errors and returns appropriate HTTP responses
func handleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	traceIDStr := getTraceID(c)

	// Check if it's already an AppError
	if appErr := errors.AsAppError(err); appErr != nil {
		logError(c, appErr, traceIDStr)

		// Build error response - minimal and clean
		errorResponse := gin.H{
			"code":    appErr.Code,
			"message": appErr.Message,
		}

		// Only include details for validation errors (user needs to know what to fix)
		// All other errors: details are logged but not exposed to client for security
		if appErr.Details != "" {
			switch appErr.Code {
			case errors.ErrorCodeValidation, errors.ErrorCodeInvalidInput, errors.ErrorCodeInvalidID:
				// Validation errors: Show details (user needs to fix their input)
				errorResponse["details"] = appErr.Details
			}
		}

		c.JSON(appErr.StatusCode, gin.H{
			"error":    errorResponse,
			"trace_id": traceIDStr,
		})
		return
	}

	// Handle standard library errors
	if err == sql.ErrNoRows {
		appErr := errors.NewNotFoundError("Resource")
		logError(c, appErr, traceIDStr)
		c.JSON(appErr.StatusCode, gin.H{
			"error": gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			},
			"trace_id": traceIDStr,
		})
		return
	}

	// Check for context timeout/cancellation
	if err == context.DeadlineExceeded {
		appErr := errors.NewRequestTimeoutError()
		logError(c, appErr, traceIDStr)
		c.JSON(appErr.StatusCode, gin.H{
			"error": gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			},
			"trace_id": traceIDStr,
		})
		return
	}

	// Default: Internal server error
	appErr := errors.NewInternalErrorWithError("An unexpected error occurred", err)
	logError(c, appErr, traceIDStr)

	// Never expose internal error details in production
	// Details are logged but not returned to client
	c.JSON(appErr.StatusCode, gin.H{
		"error": gin.H{
			"code":    appErr.Code,
			"message": "An error occurred while processing your request. Please try again later.",
		},
		"trace_id": traceIDStr,
	})
}

// logError logs the error with context information
func logError(c *gin.Context, appErr *errors.AppError, traceID string) {
	logMsg := fmt.Sprintf(
		"[ERROR] trace=%s | code=%s | status=%d | path=%s | method=%s | message=%s",
		traceID,
		appErr.Code,
		appErr.StatusCode,
		c.Request.URL.Path,
		c.Request.Method,
		appErr.Message,
	)

	if appErr.Details != "" {
		logMsg += fmt.Sprintf(" | details=%s", appErr.Details)
	}

	if appErr.Err != nil {
		logMsg += fmt.Sprintf(" | underlying_error=%v", appErr.Err)
	}

	log.Printf(logMsg)
}

// HandleError is a helper function to set an error in Gin context
// This allows services and handlers to set errors that will be handled by the middleware
func HandleError(c *gin.Context, err error) {
	if err != nil {
		c.Error(err)
		c.Abort()
	}
}

// HandleAppError is a helper function to set an AppError in Gin context
func HandleAppError(c *gin.Context, appErr *errors.AppError) {
	if appErr != nil {
		c.Error(appErr)
		c.Abort()
	}
}
