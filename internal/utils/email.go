package utils

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
)

// loadEnv loads environment variables from a .env file.
// It returns an error if the .env file cannot be loaded.
func loadEnv() error {
	return godotenv.Load()
}

// SendMail sends an email with the given subject and body to the specified recipients.
// It loads environment variables for SMTP configuration and returns an error if any step fails.
func SendMail(subject, body string, to []string) error {
	if err := loadEnv(); err != nil {
		log.Fatalf("Error loading .env file")
	}

	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return fmt.Errorf("invalid port: %v", err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("SENDER"))
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(os.Getenv("SMTP_HOST"), port, os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASS"))

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("could not send email: %v", err)
	}

	return nil
}
