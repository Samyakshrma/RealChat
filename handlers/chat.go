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
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: user_id not found in context"})
		return
	}

	userID, ok := userIDInterface.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: user_id is not an integer"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	pubsub := utils.Rdb.Subscribe(utils.Ctx, fmt.Sprintf("user:%d", userID))
	defer pubsub.Close()

	// Handle incoming messages
	go func() {
		for {
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("Error reading message:", err)
				return
			}

			if msgType == websocket.TextMessage {
				var payload map[string]interface{}
				err := json.Unmarshal(msg, &payload)
				if err != nil {
					fmt.Println("Error unmarshalling message:", err)
					continue
				}

				content := payload["content"].(string)
				createdAt := time.Now()

				// Check if it's a group message
				if groupIDRaw, exists := payload["group_id"]; exists {
					groupID := int(groupIDRaw.(float64)) // JSON numbers are float64

					// Store in DB
					query := `INSERT INTO messages (sender_id, group_id, content, created_at) VALUES ($1, $2, $3, $4)`
					_, err := config.DB.Exec(utils.Ctx, query, userID, groupID, content, createdAt)
					if err != nil {
						fmt.Println("Error saving group message:", err)
						continue
					}

					// Get all group members from DB
					rows, err := config.DB.Query(utils.Ctx,
						`SELECT user_id FROM group_members WHERE group_id = $1`, groupID)
					if err != nil {
						fmt.Println("Error fetching group members:", err)
						continue
					}
					defer rows.Close()

					// Publish to each member's Redis channel
					for rows.Next() {
						var memberID int
						rows.Scan(&memberID)
						if memberID == userID {
							continue // Don't echo back to sender
						}
						// Fetch sender's name from DB
						var senderName string
						err = config.DB.QueryRow(utils.Ctx, `SELECT username FROM users WHERE id = $1`, userID).Scan(&senderName)
						if err != nil {
							fmt.Println("Error fetching sender name:", err)
							senderName = "Unknown"
						}

						payload["sender_id"] = userID
						payload["sender_name"] = senderName
						payload["created_at"] = createdAt.Format(time.RFC3339)

						enhancedMsg, err := json.Marshal(payload)
						if err != nil {
							fmt.Println("Error marshalling enhanced group message:", err)
							continue
						}

						utils.Rdb.Publish(utils.Ctx, fmt.Sprintf("user:%d", memberID), enhancedMsg)

					}
				} else if toRaw, exists := payload["to"]; exists {
					// 1-on-1 message
					receiverID := int(toRaw.(float64))

					// Store in DB
					query := `INSERT INTO messages (sender_id, receiver_id, content, created_at) VALUES ($1, $2, $3, $4)`
					_, err := config.DB.Exec(utils.Ctx, query, userID, receiverID, content, createdAt)
					if err != nil {
						fmt.Println("Error saving DM:", err)
						continue
					}
					var senderName string
					err = config.DB.QueryRow(utils.Ctx, `SELECT username FROM users WHERE id = $1`, userID).Scan(&senderName)
					if err != nil {
						fmt.Println("Error fetching sender name:", err)
						senderName = "Unknown"
					}

					// Publish to receiver's channel
					// Inject missing fields
					payload["sender_id"] = userID
					payload["sender_name"] = senderName
					payload["created_at"] = createdAt.Format(time.RFC3339) // Send ISO string

					// Marshal payload back to JSON
					enhancedMsg, err := json.Marshal(payload)
					if err != nil {
						fmt.Println("Error marshalling enhanced message:", err)
						return
					}

					utils.Rdb.Publish(utils.Ctx, fmt.Sprintf("user:%d", receiverID), enhancedMsg)

				}
			}
		}
	}()

	// Push real-time messages from Redis to this WebSocket connection
	for {
		msg, err := pubsub.ReceiveMessage(utils.Ctx)
		if err != nil {
			fmt.Println("Error receiving pubsub message:", err)
			return
		}
		conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
	}
}

func GetDirectMessages(c *gin.Context) {
	userID := c.GetInt("user_id")
	otherID := c.Param("id")

	rows, err := config.DB.Query(utils.Ctx, `
		SELECT m.sender_id, m.receiver_id, m.content, m.created_at, u.username as sender_name
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE (m.sender_id = $1 AND m.receiver_id = $2)
		   OR (m.sender_id = $2 AND m.receiver_id = $1)
		ORDER BY m.created_at ASC
	`, userID, otherID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch direct messages"})
		return
	}
	defer rows.Close()

	messages := []gin.H{}
	for rows.Next() {
		var senderID, receiverID int
		var content, senderName string
		var createdAt time.Time
		if err := rows.Scan(&senderID, &receiverID, &content, &createdAt, &senderName); err != nil {
			continue
		}
		messages = append(messages, gin.H{
			"sender_id":   senderID,
			"sender_name": senderName,
			"receiver_id": receiverID,
			"content":     content,
			"created_at":  createdAt,
		})
	}
	c.JSON(http.StatusOK, messages)
}
