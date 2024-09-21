package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"simple-chat-app/internal/model"
	"simple-chat-app/internal/service"
)

/**
MyWebSocketServer STRUCT: A WebSocket server in Go using the gorilla/websocket package.
The server is encapsulated in the MyWebSocketServer struct, which maintains a list of connected clients,
channels for broadcasting messages, and services for handling conversations and messages.
The MyWebSocketServer struct has several fields:
clients: a map that tracks active WebSocket connections.
broadcast: a channel for broadcasting messages to all clients.
register and unregister: channels for managing client connections.
conversationService and messageService: services for handling conversation and message logic.
*/

type MyWebSocketServer struct {
	clients             map[*websocket.Conn]bool
	broadcast           chan []byte
	register            chan *websocket.Conn
	unregister          chan *websocket.Conn
	conversationService *service.ConversationService
	messageService      *service.MessageService
}

//The upgrader variable is a websocket.Upgrader that allows all origins to connect. This is used to upgrade HTTP connections to WebSocket connections.

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

/**
* The NewWebSocketServer function initializes a new instance of MyWebSocketServer, setting up the channels and services.
 */

func NewWebSocketServer(conversationService *service.ConversationService, messageService *service.MessageService) *MyWebSocketServer {
	return &MyWebSocketServer{
		clients:             make(map[*websocket.Conn]bool),
		broadcast:           make(chan []byte),
		register:            make(chan *websocket.Conn),
		unregister:          make(chan *websocket.Conn),
		conversationService: conversationService,
		messageService:      messageService,
	}
}

// The Start method starts the WebSocket server, setting up the HTTP handler and starting the message handling goroutine.
func (ws *MyWebSocketServer) Start() {
	http.HandleFunc("/ws/chat", ws.HandleConnections)
	go ws.handleMessages()
	log.Fatal(http.ListenAndServe(":8081", nil))
}

/**
* The handleMessages method listens for events on the register, unregister, and broadcast channels.
* When a new client connects, it is added to the clients map.
* When a client disconnects, it is removed from the map and the connection is closed.
* When a message is received on the broadcast channel, it is sent to all connected clients.
 */

func (ws *MyWebSocketServer) handleMessages() {
	for {
		select {
		case conn := <-ws.register:
			ws.clients[conn] = true
		case conn := <-ws.unregister:
			if _, ok := ws.clients[conn]; ok {
				delete(ws.clients, conn)
				conn.Close()
			}
		case message := <-ws.broadcast:
			for conn := range ws.clients {
				err := conn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Printf("Error writing message: %v", err)
					conn.Close()
					delete(ws.clients, conn)
				}
			}
		}
	}
}

func (ws *MyWebSocketServer) HandleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to a WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logError("Error upgrading to WebSocket", err)
		return
	}
	defer conn.Close()

	log.Printf("Client connected: %s", conn.RemoteAddr())

	// Register the connection
	ws.register <- conn

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			logError("Error reading message", err)
			ws.unregister <- conn
			break
		}

		// Broadcast the message if needed
		ws.broadcast <- message

		// Process the message
		if err := ws.processMessage(r.Context(), conn, messageType, message); err != nil {
			logError("Error processing message", err)
			continue
		}
	}
}

// processMessage handles incoming messages based on the "action" field
func (ws *MyWebSocketServer) processMessage(ctx context.Context, conn *websocket.Conn, messageType int, message []byte) error {
	var request map[string]interface{}
	if err := json.Unmarshal(message, &request); err != nil {
		return fmt.Errorf("invalid message format: %v", err)
	}

	action, ok := request["action"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid action field")
	}

	switch action {
	case "create_conversation":
		return ws.handleCreateConversation(request)

	case "get_conversationById":
		return ws.handleGetConversationById(ctx, conn, request)

	case "send_message":
		return ws.handleSendMessage(conn, request)

	default:
		log.Printf("Unknown action: %s", action)
		return nil
	}
}

// handleCreateConversation processes a request to create a new conversation
func (ws *MyWebSocketServer) handleCreateConversation(request map[string]interface{}) error {
	senderID, err := parseObjectID(request["senderId"])
	if err != nil {
		return fmt.Errorf("invalid senderId: %v", err)
	}

	receiverID, err := parseObjectID(request["receiverId"])
	if err != nil {
		return fmt.Errorf("invalid receiverId: %v", err)
	}

	conversation := model.Conversation{
		SenderId:   senderID,
		ReceiverId: receiverID,
	}

	if _, err := ws.conversationService.Create(conversation); err != nil {
		return fmt.Errorf("error creating conversation: %v", err)
	}

	log.Println("Conversation created successfully")
	return nil
}

// handleGetConversationById processes a request to retrieve a conversation by ID
func (ws *MyWebSocketServer) handleGetConversationById(ctx context.Context, conn *websocket.Conn, request map[string]interface{}) error {
	conversationID, err := parseObjectID(request["_id"])
	if err != nil {
		return fmt.Errorf("invalid conversationID: %v", err)
	}

	conversation, sender, receiver, err := ws.conversationService.GetConversationWithUsers(ctx, conversationID)
	if err != nil {
		return fmt.Errorf("error getting conversation: %v", err)
	}

	response := map[string]interface{}{
		"conversation": conversation,
		"sender":       sender,
		"receiver":     receiver,
	}
	return ws.sendResponse(conn, websocket.TextMessage, response)
}

// handleSendMessage processes a request to send a message
func (ws *MyWebSocketServer) handleSendMessage(conn *websocket.Conn, request map[string]interface{}) error {
	conversationID, err := parseObjectID(request["conversationId"])
	if err != nil {
		return fmt.Errorf("invalid conversationId: %v", err)
	}

	senderID, err := parseObjectID(request["senderId"])
	if err != nil {
		return fmt.Errorf("invalid senderId: %v", err)
	}

	message := model.Message{
		ConversationId: conversationID,
		SenderId:       senderID,
		Message:        request["message"].(string),
	}

	createdMessage, err := ws.messageService.Create(message)
	if err != nil {
		return fmt.Errorf("error creating message: %v", err)
	}

	response := map[string]interface{}{
		"status":  "success",
		"message": createdMessage,
	}

	// send Response
	return ws.sendResponse(conn, websocket.TextMessage, response)

}

// sendResponse sends a message to the WebSocket connection
func (ws *MyWebSocketServer) sendResponse(conn *websocket.Conn, messageType int, response interface{}) error {
	responseMessage, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("error marshalling response: %v", err)
	}

	if err := conn.WriteMessage(messageType, responseMessage); err != nil {
		return fmt.Errorf("error sending message: %v", err)
	}

	log.Println("Message sent successfully")
	return nil
}

// parseObjectID converts a string into a MongoDB ObjectID
func parseObjectID(id interface{}) (primitive.ObjectID, error) {
	idStr, ok := id.(string)
	if !ok {
		return primitive.NilObjectID, fmt.Errorf("id is not a string")
	}
	return primitive.ObjectIDFromHex(idStr)
}

// logError simplifies error logging
func logError(message string, err error) {
	if err != nil {
		log.Printf("%s: %v", message, err)
	}
}
