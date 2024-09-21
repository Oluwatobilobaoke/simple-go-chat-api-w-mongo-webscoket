package middleware

import (
	"net/http"
	"os"
	"simple-chat-app/internal/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

var serviceName = os.Getenv("SERVICE_NAME")

// ErrorHandlerMiddleware handles different types of errors and returns a unified response structure
func ErrorHandlerMiddleware(c *gin.Context) {
	c.Next() // Process other middlewares and handlers

	// Retrieve the last error in the gin context
	err := c.Errors.Last()
	if err == nil {
		return // No errors to handle
	}

	// Initialize default values for the response
	statusCode := http.StatusInternalServerError
	msg := "Internal Server Error"
	errorCode := "INTERNAL_SERVER_ERROR"

	// Handle different error types
	switch {
	case err.Type == gin.ErrorTypeBind:
		// Binding errors (e.g., JSON unmarshalling errors)
		statusCode = http.StatusBadRequest
		msg = "Invalid request data"
		errorCode = "BAD_REQUEST"

	case err.Type == gin.ErrorTypeRender:
		// Errors rendering the response
		statusCode = http.StatusInternalServerError
		msg = "Error rendering response"
		errorCode = "INTERNAL_SERVER_ERROR"

	case isCustomError(err):
		// Custom application-level error handling
		customErr := err.Err.(*utils.CustomError)
		statusCode = customErr.HTTPStatusCode
		msg = customErr.Message
		errorCode = "CUSTOM_ERROR"

	case isMongoError(err):
		// MongoDB-related errors
		mongoErr := err.Err.(*mongo.CommandError)
		handleMongoError(mongoErr, &statusCode, &msg, &errorCode)

	default:
		// Catch-all for other types of errors
		logError(c, err) // Optional: log the error for further analysis
	}

	// Send the error response in JSON format
	c.JSON(statusCode, gin.H{
		"success":        false,
		"message":        msg,
		"httpStatusCode": statusCode,
		"error":          errorCode,
		"service":        serviceName,
	})
}

// Helper function to identify a custom error
func isCustomError(err *gin.Error) bool {
	_, ok := err.Err.(*utils.CustomError)
	return ok
}

// Helper function to identify MongoDB errors
func isMongoError(err *gin.Error) bool {
	_, ok := err.Err.(*mongo.CommandError)
	return ok
}

// handleMongoError maps MongoDB error codes to HTTP responses
func handleMongoError(mongoErr *mongo.CommandError, statusCode *int, msg *string, errorCode *string) {
	switch mongoErr.Code {
	case 11000: // Duplicate key error
		*statusCode = http.StatusConflict
		*msg = "Duplicate value entered"
		*errorCode = "DUPLICATE_ENTRY"
	default:
		*statusCode = http.StatusInternalServerError
		*msg = "Database operation error"
		*errorCode = "DATABASE_ERROR"
	}
}

// logError can be used for logging the error details for debugging purposes
func logError(c *gin.Context, err *gin.Error) {
	// Example: You can use your preferred logging system here
	// log.Printf("Error occurred: %v, Context: %v", err.Err, c.Request)
}
