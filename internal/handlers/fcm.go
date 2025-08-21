package handlers

import (
	"net/http"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/kaelCoding/toyBE/internal/models"
	"gorm.io/gorm"
)

type FCMTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

func UpdateFCMToken(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var req FCMTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		var user models.User
		if err := db.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		user.FCMToken = req.Token
		if err := db.Save(&user).Error; err != nil {
			log.Printf("Error saving FCM token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save FCM token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "FCM token updated successfully"})
	}
}
