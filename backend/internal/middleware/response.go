// response.go - Standardized success response helpers
// Provides consistent success response format across all endpoints
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// getTraceID extracts trace ID from Gin context
func getTraceID(c *gin.Context) string {
	traceID, _ := c.Get("trace_id")
	if id, ok := traceID.(string); ok {
		return id
	}
	return ""
}

// SuccessResponse represents a standard success response
type SuccessResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	TraceID string      `json:"trace_id,omitempty"`
}

// JSONSuccess sends a standardized success response
// statusCode defaults to 200 if not provided
func JSONSuccess(c *gin.Context, data interface{}, statusCode ...int) {
	code := http.StatusOK
	if len(statusCode) > 0 && statusCode[0] > 0 {
		code = statusCode[0]
	}

	response := SuccessResponse{
		Data:    data,
		TraceID: getTraceID(c),
	}

	c.JSON(code, response)
}

// JSONSuccessWithMessage sends a standardized success response with a message
func JSONSuccessWithMessage(c *gin.Context, message string, data interface{}, statusCode ...int) {
	code := http.StatusOK
	if len(statusCode) > 0 && statusCode[0] > 0 {
		code = statusCode[0]
	}

	response := SuccessResponse{
		Data:    data,
		Message: message,
		TraceID: getTraceID(c),
	}

	c.JSON(code, response)
}

// JSONCreated sends a 201 Created response (for POST requests)
func JSONCreated(c *gin.Context, data interface{}) {
	JSONSuccess(c, data, http.StatusCreated)
}

// JSONNoContent sends a 204 No Content response
func JSONNoContent(c *gin.Context) {
	c.JSON(http.StatusNoContent, gin.H{
		"trace_id": getTraceID(c),
	})
}
