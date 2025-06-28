package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/Samyakshrma/RealChat/config"
	"github.com/Samyakshrma/RealChat/handlers"
	"github.com/Samyakshrma/RealChat/middleware"
	"github.com/Samyakshrma/RealChat/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	config.InitDB(os.Getenv("DATABASE_URL"))
	ctx := context.Background()
	utils.InitRedis(ctx)

	// Trust local proxy only (avoid security warning in Gin)
	//gin.SetTrustedProxies([]string{"127.0.0.1"})

	r := gin.Default()

	// Allow CORS from frontend (e.g., Next.js on port 3000)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Auth routes
	r.POST("/login", handlers.Login)
	r.POST("/register", handlers.Register)

	// WebSocket chat
	r.GET("/chat", middleware.AuthMiddleware(), handlers.ChatHandler)

	// Group chat routes
	r.POST("/groups", middleware.AuthMiddleware(), handlers.CreateGroup)
	r.GET("/groups", middleware.AuthMiddleware(), handlers.GetUserGroups)
	r.GET("/groups/:id/messages", middleware.AuthMiddleware(), handlers.GetGroupMessages)
	r.POST("/groups/:id/add-member", middleware.AuthMiddleware(), handlers.AddGroupMember)

	// Direct messages
	r.GET("/users/:id/messages", middleware.AuthMiddleware(), handlers.GetDirectMessages)

	// Start server
	r.Run(":8080")
}
