package handlers

import (
	"log"
	"net/http"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/kaelCoding/toyBE/internal/chat"
	"github.com/kaelCoding/toyBE/internal/models"
	"gorm.io/gorm"
)

func ChatEndpoint(hub *chat.Hub, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var user models.User
		if err := db.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		log.Printf("Setting up WebSocket for user ID: %d, IsAdmin: %v", user.ID, user.Admin)
		chat.ServeWs(hub, c, user.ID, user.Admin)
	}
}

func GetChatHistory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var user models.User
		if err := db.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		var messages []models.Message
		var queryUserID uint

		if user.Admin {
			otherUserID := c.Query("userId")
			if otherUserID == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "userId query parameter is required for admin"})
				return
			}
			var id uint
			_, err := fmt.Sscan(otherUserID, &id)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userId"})
				return
			}
			queryUserID = id
		} else {
			queryUserID = user.ID
		}

		var adminUser models.User
		db.Where("admin = ?", true).First(&adminUser)
		adminID := adminUser.ID

		err := db.Preload("Sender").Preload("Receiver").
			Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
				queryUserID, adminID, adminID, queryUserID).
			Order("timestamp asc").Find(&messages).Error

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve messages"})
			return
		}

		c.JSON(http.StatusOK, messages)
	}
}
