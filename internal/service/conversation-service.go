package service

import (
	"context"
	"errors"
	"simple-chat-app/internal/model"
	"simple-chat-app/internal/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ConversationService provides methods to manage conversations.
type ConversationService struct {
	conversationCollection *mongo.Collection
	userCollection         *mongo.Collection
}

// NewConversationService creates a new ConversationService with the given database.
func NewConversationService(db *mongo.Database) *ConversationService {
	return &ConversationService{
		conversationCollection: db.Collection("conversation"),
		userCollection:         db.Collection("user"),
	}
}

// validateUserInput checks if the conversation has valid sender and receiver IDs.
// Returns an error if either ID is missing.
func (cs *ConversationService) validateUserInput(conversation model.Conversation) error {
	if conversation.SenderId == primitive.NilObjectID || conversation.ReceiverId == primitive.NilObjectID {
		return errors.New("senderId and receiverId are required")
	}
	return nil
}

// Create adds a new conversation to the database if it is valid and does not already exist.
// Returns the created conversation or an error if the operation fails.
func (cs *ConversationService) Create(conversation model.Conversation) (*model.Conversation, error) {
	if err := cs.validateUserInput(conversation); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if both sender and receiver exist
	userFilter := bson.M{
		"_id": bson.M{"$in": []primitive.ObjectID{conversation.SenderId, conversation.ReceiverId}},
	}

	count, err := cs.userCollection.CountDocuments(ctx, userFilter)
	if err != nil || count < 2 {
		return nil, utils.NewBadRequestError("One or both users do not exist")
	}

	// Check if conversation exists
	conversationFilter := bson.M{
		"senderId":   conversation.SenderId,
		"receiverId": conversation.ReceiverId,
	}

	var existingConv model.Conversation
	err = cs.conversationCollection.FindOne(ctx, conversationFilter).Decode(&existingConv)
	if err == nil {
		return nil, utils.NewConflictError("Conversation already exists")
	}
	if err != mongo.ErrNoDocuments {
		return nil, errors.New("internal server error")
	}

	conversation.ID = primitive.NewObjectID()
	conversation.CreatedAt = time.Now()
	conversation.UpdatedAt = time.Now()

	_, err = cs.conversationCollection.InsertOne(ctx, conversation)
	if err != nil {
		return nil, err
	}

	return &conversation, nil
}

// GetConversationWithUsers retrieves a conversation and its associated users by conversation ID.
// Returns the conversation, sender, receiver, and an error if the operation fails.
func (cs *ConversationService) GetConversationWithUsers(ctx context.Context, convID primitive.ObjectID) (*model.Conversation, *model.User, *model.User, error) {
	var conversation model.Conversation
	if err := cs.conversationCollection.FindOne(ctx, bson.M{"_id": convID}).Decode(&conversation); err != nil {
		return nil, nil, nil, err
	}

	var sender, receiver model.User
	if err := cs.userCollection.FindOne(ctx, bson.M{"_id": conversation.SenderId}).Decode(&sender); err != nil {
		return &conversation, nil, nil, err
	}
	if err := cs.userCollection.FindOne(ctx, bson.M{"_id": conversation.ReceiverId}).Decode(&receiver); err != nil {
		return &conversation, &sender, nil, err
	}

	return &conversation, &sender, &receiver, nil
}
