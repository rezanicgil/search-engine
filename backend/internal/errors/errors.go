// errors.go - Custom error types and error handling utilities
// Provides structured error handling with HTTP status codes and error codes
package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents a specific error type
type ErrorCode string

const (
	// Validation errors (400)
	ErrorCodeValidation   ErrorCode = "VALIDATION_ERROR"
	ErrorCodeInvalidInput ErrorCode = "INVALID_INPUT"
	ErrorCodeInvalidID    ErrorCode = "INVALID_ID"

	// Not found errors (404)
	ErrorCodeNotFound         ErrorCode = "NOT_FOUND"
	ErrorCodeContentNotFound  ErrorCode = "CONTENT_NOT_FOUND"
	ErrorCodeProviderNotFound ErrorCode = "PROVIDER_NOT_FOUND"

	// Timeout errors (408, 504)
	ErrorCodeTimeout        ErrorCode = "TIMEOUT"
	ErrorCodeRequestTimeout ErrorCode = "REQUEST_TIMEOUT"
	ErrorCodeQueryTimeout   ErrorCode = "QUERY_TIMEOUT"

	// Internal server errors (500)
	ErrorCodeInternal ErrorCode = "INTERNAL_ERROR"
	ErrorCodeDatabase ErrorCode = "DATABASE_ERROR"
	ErrorCodeCache    ErrorCode = "CACHE_ERROR"
	ErrorCodeService  ErrorCode = "SERVICE_ERROR"

	// Service unavailable (503)
	ErrorCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
)

// AppError represents an application error with structured information
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	StatusCode int       `json:"-"`
	Err        error     `json:"-"` // Original error for logging
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new AppError
func NewAppError(code ErrorCode, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// NewAppErrorWithDetails creates a new AppError with details
func NewAppErrorWithDetails(code ErrorCode, message, details string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Details:    details,
		StatusCode: statusCode,
	}
}

// NewAppErrorWithError creates a new AppError wrapping an existing error
func NewAppErrorWithError(code ErrorCode, message string, statusCode int, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
		Details:    err.Error(),
	}
}

// Predefined error constructors for common cases

// NewValidationError creates a validation error
func NewValidationError(message string) *AppError {
	return NewAppError(ErrorCodeValidation, message, http.StatusBadRequest)
}

// NewValidationErrorWithDetails creates a validation error with details
func NewValidationErrorWithDetails(message, details string) *AppError {
	return NewAppErrorWithDetails(ErrorCodeValidation, message, details, http.StatusBadRequest)
}

// NewInvalidInputError creates an invalid input error
func NewInvalidInputError(message string) *AppError {
	return NewAppError(ErrorCodeInvalidInput, message, http.StatusBadRequest)
}

// NewInvalidIDError creates an invalid ID error
func NewInvalidIDError(resource string) *AppError {
	return NewAppError(ErrorCodeInvalidID, fmt.Sprintf("Invalid %s ID", resource), http.StatusBadRequest)
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *AppError {
	return NewAppError(ErrorCodeNotFound, fmt.Sprintf("%s not found", resource), http.StatusNotFound)
}

// NewContentNotFoundError creates a content not found error
func NewContentNotFoundError() *AppError {
	return NewAppError(ErrorCodeContentNotFound, "Content not found", http.StatusNotFound)
}

// NewContentNotFoundErrorWithID creates a content not found error with content ID
func NewContentNotFoundErrorWithID(id int64) *AppError {
	return NewAppErrorWithDetails(
		ErrorCodeContentNotFound,
		"Content not found",
		fmt.Sprintf("No content found with ID: %d", id),
		http.StatusNotFound,
	)
}

// NewProviderNotFoundError creates a provider not found error
func NewProviderNotFoundError() *AppError {
	return NewAppError(ErrorCodeProviderNotFound, "Provider not found", http.StatusNotFound)
}

// NewProviderNotFoundErrorWithName creates a provider not found error with provider name
func NewProviderNotFoundErrorWithName(name string) *AppError {
	return NewAppErrorWithDetails(
		ErrorCodeProviderNotFound,
		"Provider not found",
		fmt.Sprintf("No provider found with name: %s", name),
		http.StatusNotFound,
	)
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(message string) *AppError {
	return NewAppError(ErrorCodeTimeout, message, http.StatusRequestTimeout)
}

// NewRequestTimeoutError creates a request timeout error
func NewRequestTimeoutError() *AppError {
	return NewAppError(ErrorCodeRequestTimeout, "Request timeout", http.StatusRequestTimeout)
}

// NewRequestTimeoutErrorWithDuration creates a request timeout error with duration
func NewRequestTimeoutErrorWithDuration(duration string) *AppError {
	return NewAppErrorWithDetails(
		ErrorCodeRequestTimeout,
		"Request timeout",
		fmt.Sprintf("The request exceeded the maximum allowed time (%s)", duration),
		http.StatusRequestTimeout,
	)
}

// NewQueryTimeoutError creates a query timeout error
func NewQueryTimeoutError(operation string) *AppError {
	return NewAppErrorWithDetails(
		ErrorCodeQueryTimeout,
		fmt.Sprintf("Query timeout: %s", operation),
		fmt.Sprintf("The %s operation exceeded the maximum allowed time", operation),
		http.StatusGatewayTimeout,
	)
}

// NewInternalError creates an internal server error
func NewInternalError(message string) *AppError {
	return NewAppError(ErrorCodeInternal, message, http.StatusInternalServerError)
}

// NewInternalErrorWithError creates an internal server error wrapping an existing error
func NewInternalErrorWithError(message string, err error) *AppError {
	return NewAppErrorWithError(ErrorCodeInternal, message, http.StatusInternalServerError, err)
}

// NewDatabaseError creates a database error
func NewDatabaseError(operation string, err error) *AppError {
	return NewAppErrorWithError(
		ErrorCodeDatabase,
		fmt.Sprintf("Database operation failed: %s", operation),
		http.StatusInternalServerError,
		err,
	)
}

// NewServiceError creates a service error
func NewServiceError(operation string, err error) *AppError {
	return NewAppErrorWithError(
		ErrorCodeService,
		fmt.Sprintf("Service operation failed: %s", operation),
		http.StatusInternalServerError,
		err,
	)
}

// NewServiceUnavailableError creates a service unavailable error
func NewServiceUnavailableError(message string) *AppError {
	return NewAppError(ErrorCodeServiceUnavailable, message, http.StatusServiceUnavailable)
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// AsAppError converts an error to AppError if possible
func AsAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return nil
}

// WrapError wraps a standard error into an AppError
func WrapError(err error, code ErrorCode, message string, statusCode int) *AppError {
	if err == nil {
		return nil
	}

	// If it's already an AppError, return it
	if appErr := AsAppError(err); appErr != nil {
		return appErr
	}

	return NewAppErrorWithError(code, message, statusCode, err)
}

// Standard error variables for backward compatibility
var (
	ErrContentNotFound  = NewContentNotFoundError()
	ErrProviderNotFound = NewProviderNotFoundError()
)
