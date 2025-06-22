package main

import (
	"context"
	"log"
	"os"

	"github.com/Samyakshrma/RealChat/config"
	"github.com/Samyakshrma/RealChat/handlers"
	"github.com/Samyakshrma/RealChat/middleware"
	"github.com/Samyakshrma/RealChat/utils"
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

	r := gin.Default()
	r.POST("/login", handlers.Login)
	r.POST("/register", handlers.Register)
	r.GET("/chat", middleware.AuthMiddleware(), handlers.ChatHandler)

	r.Run(":8080")
}
