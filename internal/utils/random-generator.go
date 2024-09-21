package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

// GenerateRandomNumber generates a random number between 100000 and 999999.
// It returns the generated number as a string and an error if any occurs during the generation.
func GenerateRandomNumber() (string, error) {
	maxValue := big.NewInt(900000) // 999999 - 100000 + 1
	n, err := rand.Int(rand.Reader, maxValue)
	if err != nil {
		return "", fmt.Errorf("error generating random number: %v", err)
	}
	n.Add(n, big.NewInt(100000))
	return n.String(), nil
}

// GetOtpExpiryTime returns the current time plus 10 minutes.
// This function is used to determine the expiration time for an OTP (One-Time Password).
func GetOtpExpiryTime() time.Time {
	return time.Now().Add(10 * time.Minute)
}
