package service

import (
	"context"
	"errors"
	"simple-chat-app/internal/model"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// MessageService provides methods to manage messages.
type MessageService struct {
	conversationCollection *mongo.Collection
	messageCollection      *mongo.Collection
}

// NewMessageService creates a new MessageService with the given database.
func NewMessageService(db *mongo.Database) *MessageService {
	return &MessageService{
		conversationCollection: db.Collection("conversation"),
		messageCollection:      db.Collection("message"),
	}
}

// validateUserInput checks if the message has valid sender and conversation IDs.
// Returns an error if either ID is missing.
func (ms *MessageService) validateUserInput(message model.Message) error {
	if message.SenderId == primitive.NilObjectID || message.ConversationId == primitive.NilObjectID {
		return errors.New("ConversationId and SenderId are required")
	}
	return nil
}

// Create adds a new message to the database if it is valid.
// Returns the created message or an error if the operation fails.
func (ms *MessageService) Create(message model.Message) (*model.Message, error) {

	if err := ms.validateUserInput(message); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	message.ID = primitive.NewObjectID()
	message.CreatedAt = time.Now()
	message.UpdatedAt = time.Now()

	_, err := ms.messageCollection.InsertOne(ctx, message)
	if err != nil {
		return nil, err
	}

	return &message, nil
}
