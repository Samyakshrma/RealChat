package handlers

import (
	"net/http"
	"time"

	"github.com/Samyakshrma/RealChat/config"
	"github.com/Samyakshrma/RealChat/utils"
	"github.com/gin-gonic/gin"
)

// Struct for group creation
type CreateGroupRequest struct {
	Name      string `json:"name"`
	MemberIDs []int  `json:"member_ids"`
}

func CreateGroup(c *gin.Context) {
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	creatorID := c.GetInt("user_id")
	tx, err := config.DB.Begin(utils.Ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group"})
		return
	}
	defer tx.Rollback(utils.Ctx)

	var groupID int
	err = tx.QueryRow(utils.Ctx, `INSERT INTO groups (name, created_by, created_at) VALUES ($1, $2, $3) RETURNING id`,
		req.Name, creatorID, time.Now()).Scan(&groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not insert group"})
		return
	}

	// Insert creator and members
	allMembers := append(req.MemberIDs, creatorID)
	for _, uid := range allMembers {
		_, err := tx.Exec(utils.Ctx, `INSERT INTO group_members (group_id, user_id) VALUES ($1, $2)`, groupID, uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add group members"})
			return
		}
	}

	tx.Commit(utils.Ctx)
	c.JSON(http.StatusOK, gin.H{"group_id": groupID, "message": "Group created successfully"})
}

func GetUserGroups(c *gin.Context) {
	userID := c.GetInt("user_id")

	rows, err := config.DB.Query(utils.Ctx, `
		SELECT g.id, g.name, g.created_at
		FROM groups g
		JOIN group_members gm ON gm.group_id = g.id
		WHERE gm.user_id = $1
	`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch groups"})
		return
	}
	defer rows.Close()

	groups := []gin.H{}
	for rows.Next() {
		var id int
		var name string
		var createdAt time.Time
		if err := rows.Scan(&id, &name, &createdAt); err != nil {
			continue
		}
		groups = append(groups, gin.H{
			"id":         id,
			"name":       name,
			"created_at": createdAt,
		})
	}
	c.JSON(http.StatusOK, groups)
}

func GetGroupMessages(c *gin.Context) {
	groupID := c.Param("id")

	rows, err := config.DB.Query(utils.Ctx, `
		SELECT sender_id, content, created_at
		FROM messages
		WHERE group_id = $1
		ORDER BY created_at ASC
	`, groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}
	defer rows.Close()

	messages := []gin.H{}
	for rows.Next() {
		var senderID int
		var content string
		var createdAt time.Time
		if err := rows.Scan(&senderID, &content, &createdAt); err != nil {
			continue
		}
		messages = append(messages, gin.H{
			"sender_id":  senderID,
			"content":    content,
			"created_at": createdAt,
		})
	}
	c.JSON(http.StatusOK, messages)
}

func AddGroupMember(c *gin.Context) {
	groupID := c.Param("id")
	var body struct {
		UserID int `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	_, err := config.DB.Exec(utils.Ctx,
		`INSERT INTO group_members (group_id, user_id) VALUES ($1, $2)`,
		groupID, body.UserID,
	)
	if err != nil {
		if pgErr, ok := err.(interface{ SQLState() string }); ok && pgErr.SQLState() == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "User already in group"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User added to group"})
}
