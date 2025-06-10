package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/Samyakshrma/RealChat/config"
	"github.com/Samyakshrma/RealChat/models"
)

func Login(c *gin.Context) {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	c.BindJSON(&creds)

	var user models.User
	err := config.DB.QueryRow(context.Background(),
		"SELECT id, password_hash FROM users WHERE username=$1", creds.Username).
		Scan(&user.ID, &user.PasswordHash)

	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(creds.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	})
	tokenStr, _ := token.SignedString([]byte("SECRET"))
	c.JSON(200, gin.H{"token": tokenStr})
}
