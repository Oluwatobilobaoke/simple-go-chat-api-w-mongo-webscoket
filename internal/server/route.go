package server

import (
	"github.com/gin-gonic/gin"
	"os"
	"simple-chat-app/internal/controller"
	"simple-chat-app/internal/middleware"

	"net/http"
)

func (s *Server) RegisterRoutes() http.Handler {

	// Load environment variables (ideally done once at startup)
	_ = os.Getenv("JWT_SECRET")
	_ = os.Getenv("SERVICE_NAME")
	r := gin.Default()

	r.Use(gin.Logger())

	r.GET("/", s.HelloWorldHandler)

	r.Use(middleware.ErrorHandlerMiddleware)
	r.NoRoute(middleware.HandleNotFound)

	userController := controller.NewUserController()

	r.POST("/v1/auth/users/create", userController.CreateUserHttp)
	r.POST("/v1/auth/users/verify-email", userController.VerifyEmailHandler)

	r.POST("/v1/auth/users/send-email", userController.SendEmailHandler)

	r.POST("/v1/auth/users/login", userController.LoginHandler)

	// Apply the middleware to your routes
	authorized := r.Group("/v1/auth")
	authorized.Use(middleware.VerifyToken(os.Getenv("SERVICE_NAME"), os.Getenv("JWT_SECRET")))

	//{
	//	authorized.POST("/users/upload-image", userController.UploadImageHandler)
	//}

	r.GET("/ws", func(c *gin.Context) {
		s.ws.HandleConnections(c.Writer, c.Request)
	})

	return r
}

func (s *Server) HelloWorldHandler(c *gin.Context) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	c.JSON(http.StatusOK, resp)
}
