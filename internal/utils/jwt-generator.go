package utils

import (
	"github.com/golang-jwt/jwt/v5"
	"os"
	"time"
)

var myJwtSigningKey = []byte(os.Getenv("JWT_SECRET"))

type MyClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateJWT generates a new JWT with the given user ID.
// It returns the JWT as a string.
func GenerateJWT(userID string) (string, error) {
	claims := MyClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenGenerated, err := token.SignedString(myJwtSigningKey)
	if err != nil {
		// return NewBadRequestError error
		return "", err
	}
	return tokenGenerated, nil
}

// ParseJWT parses the given JWT and returns the claims.
// If the JWT is invalid, the function returns an error.
func ParseJWT(tokenString string) (*MyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return myJwtSigningKey, nil
	})
	if err != nil {
		// return NewBadRequestError error
		return nil, err
	}
	claims, ok := token.Claims.(*MyClaims)
	if !ok {
		// return NewBadRequestError error
		return nil, err
	}
	return claims, nil
}
