package common

import (
	"errors"
	"regexp"
)

type ValidationResult struct {
	IsValid bool
}

var (
	minLengthError = errors.New("password must be at least 8 characters long")
	maxLengthError = errors.New("password must be no more than 20 characters long")
	digitError     = errors.New("password must contain at least one digit")
	lowercaseError = errors.New("password must contain at least one lowercase letter")
	uppercaseError = errors.New("password must contain at least one uppercase letter")

	// Precompiled regular expressions for better performance
	hasDigit = regexp.MustCompile(`[0-9]`)
	hasLower = regexp.MustCompile(`[a-z]`)
	hasUpper = regexp.MustCompile(`[A-Z]`)
)

func ValidatePassword(password string) (*ValidationResult, error) {
	// Length checks
	if len(password) < 8 {
		return &ValidationResult{IsValid: false}, minLengthError
	}
	if len(password) > 20 {
		return &ValidationResult{IsValid: false}, maxLengthError
	}

	// Pattern checks
	if !hasDigit.MatchString(password) {
		return &ValidationResult{IsValid: false}, digitError
	}
	if !hasLower.MatchString(password) {
		return &ValidationResult{IsValid: false}, lowercaseError
	}
	if !hasUpper.MatchString(password) {
		return &ValidationResult{IsValid: false}, uppercaseError
	}

	// If all checks pass, the password is valid
	return &ValidationResult{IsValid: true}, nil
}
