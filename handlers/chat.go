package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Samyakshrma/RealChat/config"
	"github.com/Samyakshrma/RealChat/utils"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func ChatHandler(c *gin.Context) {
	userID := c.GetInt("user_id")
	targetID := c.Query("to")

	fmt.Printf("User %d connected to chat with user %s\n", userID, targetID)

	conn, _ := upgrader.Upgrade(c.Writer, c.Request, nil)
	pubsub := utils.Rdb.Subscribe(utils.Ctx, fmt.Sprintf("user:%s", targetID))
	defer pubsub.Close()

	go func() {
		for {
			msgType, msg, _ := conn.ReadMessage()
			if msgType == websocket.TextMessage {
				var payload map[string]interface{}
				json.Unmarshal(msg, &payload)

				// Save to DB (Use config.DB instead of utils.Db)
				query := `INSERT INTO messages (sender_id, receiver_id, content, created_at) VALUES ($1, $2, $3, $4)`
				_, err := config.DB.Exec(utils.Ctx, query, userID, payload["to"], payload["content"], time.Now())
				if err != nil {
					fmt.Println("Error saving message:", err)
				}

				// Publish to receiver
				utils.Rdb.Publish(utils.Ctx, fmt.Sprintf("user:%s", payload["to"]), msg)
			}
		}
	}()

	for {
		msg, _ := pubsub.ReceiveMessage(utils.Ctx)
		conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
	}
}
