package service

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"simple-chat-app/internal/common"
	"simple-chat-app/internal/database"
	"simple-chat-app/internal/model"
	"simple-chat-app/internal/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserService struct {
	collection *mongo.Collection
}

func NewUserService() *UserService {
	client, err := database.DBinstance()
	if err != nil {
		panic(err)
	}
	return &UserService{
		collection: client.Database("Gomongodb").Collection("user"),
	}
}

func (s *UserService) validateUserInput(user model.User) error {
	if user.Email == "" || user.Username == "" || user.Password == "" {
		return errors.New("email, username, and password are required")
	}
	return nil
}

func (s *UserService) Create(user model.User) (*model.User, error) {
	if err := s.validateUserInput(user); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if user email or username already exists
	filter := bson.M{
		"$or": []bson.M{
			{"email": user.Email},
			{"username": user.Username},
		},
	}

	var existingUser model.User
	err := s.collection.FindOne(ctx, filter).Decode(&existingUser)
	if err == nil {
		return nil, utils.NewConflictError("user with given email or username already exists")
	}
	if err != mongo.ErrNoDocuments {
		return nil, errors.New("internal server error")
	}

	// Validate password
	if passwordValidation, err := common.ValidatePassword(user.Password); err != nil || !passwordValidation.IsValid {
		return nil, errors.New("password is not valid")
	}

	user.Password = utils.HashPassword(user.Password)
	user.VerifiedEmail = false

	otpToken, err := utils.GenerateRandomNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate OTP token: %w", err)
	}
	user.OtpToken = otpToken
	user.ExpiredAt = utils.GetOtpExpiryTime()
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	fmt.Println(otpToken)

	go func() {
		to := []string{user.Email}
		subject := "Test Email"
		body := fmt.Sprintf("<h1>Hello from YDM GO Server! Here is your OTP: %s</h1>", otpToken)

		if err := utils.SendMail(subject, body, to); err != nil {
			log.Printf("Could not send email: %v", err)
		} else {
			fmt.Println("Email sent successfully!")
		}
	}()

	if _, err = s.collection.InsertOne(ctx, user); err != nil {
		return nil, err
	}

	return &user, nil
}

// VerifyEmail verifies a user's email using the provided OTP token.
// It checks if the user exists, if the OTP token matches, and if the token is not expired.
// If the verification is successful, it updates the user's email verification status.
//
// Parameters:
//   - email: The email address of the user to verify.
//   - otpToken: The OTP token provided by the user for verification.
//
// Returns:
//   - error: An error if the verification fails, or nil if it succeeds.
func (s *UserService) VerifyEmail(email string, otpToken string) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find user by email
	var user model.User
	err := s.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("user not found")
		}
		return fmt.Errorf("error finding user: %w", err)
	}

	// Check if OTP token matches and is not expired
	if user.OtpToken != otpToken {
		return errors.New("invalid OTP token")
	}
	if user.ExpiredAt.Before(time.Now()) {
		return errors.New("OTP token has expired")
	}

	update := bson.M{
		"$set": bson.M{
			"verifiedEmail": true,
			"otpToken":      nil,
			"expiredAt":     nil,
		},
	}
	_, err = s.collection.UpdateOne(ctx, bson.M{"email": email}, update)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (s *UserService) SendMail(email string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if user with the given email exists
	filter := bson.M{"email": email}
	var existingUser model.User
	err := s.collection.FindOne(ctx, filter).Decode(&existingUser)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", errors.New("user does not exist")
		}
		return "", fmt.Errorf("error finding user: %w", err)
	}

	// Generate OTP token
	otpToken, err := utils.GenerateRandomNumber()
	if err != nil {
		return "", fmt.Errorf("failed to generate OTP token: %w", err)
	}

	// Update user's OTP token and expiry time
	update := bson.M{
		"$set": bson.M{
			"otpToken":  otpToken,
			"expiredAt": utils.GetOtpExpiryTime(),
		},
	}
	_, err = s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return "", fmt.Errorf("failed to update user: %w", err)
	}

	// Send email with OTP token
	to := []string{email}
	subject := "OTP for account verification"
	body := fmt.Sprintf("Your OTP is: %s", otpToken)

	if err := utils.SendMail(subject, body, to); err != nil {
		return "", fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Email sent successfully to %s\n", email)

	// Return success message
	successMessage := fmt.Sprintf("Email sent successfully to %s", email)
	return successMessage, nil
}

// Login authenticates a user by verifying their email and password.
// It checks if the user exists, if the password is correct, and if the email is verified.
// If the authentication is successful, it generates and returns a JWT token.
//
// Parameters:
//   - email: The email address of the user attempting to log in.
//   - password: The password provided by the user for authentication.
//
// Returns:
//   - string: A JWT token if the authentication is successful.
//   - error: An error if the authentication fails.
func (s *UserService) Login(email, password string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user model.User
	err := s.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", errors.New("user not found")
		}
		return "", errors.New("internal server error")
	}

	// Check if password is correct
	if !utils.VerifyPassword(password, user.Password) {
		return "", errors.New("invalid password")
	}

	if !user.VerifiedEmail {
		return "", errors.New("please verify your email")
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID.Hex())
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return token, nil
}
