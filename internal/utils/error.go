package utils

import (
	"encoding/json"
	"net/http"
	"os"
)

// CustomError represents a custom error with additional fields for error code, HTTP status code, service name, and success status.
type CustomError struct {
	Message        string `json:"message"`
	ErrorCode      int    `json:"errorCode,omitempty"`
	HTTPStatusCode int    `json:"httpStatusCode,omitempty"`
	Service        string `json:"service,omitempty"`
	Success        bool   `json:"success,omitempty"`
}

// Error returns the error message of the CustomError.
func (e *CustomError) Error() string {
	return e.Message
}

var serviceName = os.Getenv("SERVICE_NAME")

// newError creates a new CustomError with the given message, error code, and HTTP status code.
func newError(message string, errorCode, httpStatusCode int) *CustomError {
	return &CustomError{
		Message:        message,
		ErrorCode:      errorCode,
		HTTPStatusCode: httpStatusCode,
		Service:        serviceName,
		Success:        false,
	}
}

// NewUnauthorizedError creates a new CustomError for unauthorized access (HTTP 401).
func NewUnauthorizedError(message string) *CustomError {
	return newError(message, 401, http.StatusUnauthorized)
}

// NewBadRequestError creates a new CustomError for bad requests (HTTP 400).
func NewBadRequestError(message string) *CustomError {
	return newError(message, 400, http.StatusBadRequest)
}

// NewConflictError creates a new CustomError for conflicts (HTTP 409).
func NewConflictError(message string) *CustomError {
	return newError(message, 409, http.StatusConflict)
}

// NewInternalServerError creates a new CustomError for internal server errors (HTTP 500).
func NewInternalServerError(message string) *CustomError {
	return newError(message, 500, http.StatusInternalServerError)
}

// NewUnauthenticatedError creates a new CustomError for unauthenticated access (HTTP 401).
func NewUnauthenticatedError(message string) *CustomError {
	return newError(message, 401, http.StatusUnauthorized)
}

// NewNotFoundError creates a new CustomError for not found errors (HTTP 404).
func NewNotFoundError(message string) *CustomError {
	return newError(message, 404, http.StatusNotFound)
}

// ToJSON converts the CustomError to a JSON byte slice.
func (e *CustomError) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}
