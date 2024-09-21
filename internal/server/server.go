package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/mongo"

	"simple-chat-app/internal/database"
	"simple-chat-app/internal/service"
	"simple-chat-app/internal/websocket"
)

type Server struct {
	port int
	db   *mongo.Database
	ws   *websocket.MyWebSocketServer
}

func NewServer() *http.Server {
	portStr := os.Getenv("PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		port = 8080
	}

	db, err := database.New()
	if err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		os.Exit(1)
	}

	conversationService := service.NewConversationService(db)
	messageService := service.NewMessageService(db)
	ws := websocket.NewWebSocketServer(conversationService, messageService)
	go ws.Start()

	newServer := &Server{
		port: port,
		db:   db,
		ws:   ws,
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", newServer.port),
		Handler:      newServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
