package utils

import (
	"golang.org/x/crypto/bcrypt"
	"log"
)

// HashPassword generates a bcrypt hash of the given password.
// It returns the hashed password as a string.
// If an error occurs during hashing, the function logs the error and panics.
func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}

	return string(bytes)
}

// HashPasswordReturnsHash generates a bcrypt hash of the given password.
// It returns the hashed password as a string.
// If an error occurs during hashing, the function logs the error and panics.
func HashPasswordReturnsHash(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(hash)
}

// VerifyPassword compares a bcrypt hashed password with its possible plaintext equivalent.
// It returns true if the passwords match, and false otherwise.
func VerifyPassword(providedPassword, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(providedPassword))
	return err == nil
}
