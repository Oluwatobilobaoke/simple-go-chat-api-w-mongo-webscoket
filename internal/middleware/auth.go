package middleware

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ErrorResponse is a helper function for sending a JSON error response
func ErrorResponse(c *gin.Context, statusCode int, message, errorCode, serviceName string) {
	c.JSON(statusCode, gin.H{
		"success":        false,
		"message":        message,
		"httpStatusCode": statusCode,
		"error":          errorCode,
		"service":        serviceName,
	})
	c.Abort()
}

// VerifyToken middleware ensures that a valid JWT token is present in the Authorization header
func VerifyToken(serviceName string, accessTokenSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			ErrorResponse(c, http.StatusUnauthorized, "Authorization header is missing or invalid", "VALIDATION_ERROR", serviceName)
			return
		}

		// Extract token from the Authorization header
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse and verify the JWT token
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(accessTokenSecret), nil
		})

		if err != nil || !token.Valid {
			ErrorResponse(c, http.StatusUnauthorized, "Invalid token", "VALIDATION_ERROR", serviceName)
			return
		}

		// Extract user ID from the token claims
		userID, err := extractUserIDFromClaims(claims)
		if err != nil {
			ErrorResponse(c, http.StatusUnauthorized, err.Error(), "VALIDATION_ERROR", serviceName)
			return
		}

		// Set the user ID in the context for the next handlers
		c.Set("userID", userID)

		// Proceed to the next middleware or handler
		c.Next()
	}
}

// extractUserIDFromClaims extracts the user ID from JWT claims with proper type assertion
func extractUserIDFromClaims(claims jwt.MapClaims) (string, error) {
	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", errors.New("invalid token format: user_id is missing or invalid")
	}
	return userID, nil
}
