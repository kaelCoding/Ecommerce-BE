package handlers

import (
	"log"
	"net/http"
	"time"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kaelCoding/toyBE/internal/models"
	"github.com/kaelCoding/toyBE/internal/services"
	"gorm.io/gorm"
)

func SpinByShippingCode(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.SpinRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
			return
		}

		var order models.Order
		if err := db.Where("shipping_code = ?", req.ShippingCode).First(&order).Error; err != nil {
			log.Printf("Failed to find order with shipping code %s: %v", req.ShippingCode, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid or expired shipping code."})
			return
		}

		if order.HasSpun {
			c.JSON(http.StatusConflict, gin.H{"error": "This shipping code has already been used to spin."})
			return
		}

		tx := db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "An internal error occurred."})
			}
		}()

		reward, err := services.SpinWheel(tx)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to spin the wheel."})
			return
		}

		order.HasSpun = true
		if err := tx.Save(&order).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status."})
			return
		}

		spinLog := models.SpinLog{
			OrderID:  order.ID,
			UserID:   order.UserID,
			RewardID: reward.ID,
			SpinDate: time.Now(),
		}
		if err := tx.Create(&spinLog).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log spin result."})
			return
		}

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction."})
			return
		}

		c.JSON(http.StatusOK, models.SpinResponse{
			Message: "Congratulations! You won a reward.",
			Reward:  *reward,
		})
	}
}

func UpdateShippingCode(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("id")
		var req struct {
			ShippingCode string `json:"shippingCode" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		var order models.Order
		if err := db.First(&order, orderID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}

		order.ShippingCode = req.ShippingCode
		if err := db.Save(&order).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update shipping code"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Shipping code updated successfully"})
	}
}

func GetRewards(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var rewards []models.Reward
		if err := db.Find(&rewards).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve rewards"})
			return
		}
		c.JSON(http.StatusOK, rewards)
	}
}

func AddReward(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var newReward models.Reward
		if err := c.ShouldBindJSON(&newReward); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reward data: " + err.Error()})
			return
		}

		if err := db.Create(&newReward).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add new reward: " + err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Reward added successfully", "reward": newReward})
	}
}

func UpdateReward(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rewardID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reward ID"})
			return
		}

		var existingReward models.Reward
		if err := db.First(&existingReward, uint(rewardID)).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Reward not found"})
			return
		}

		var updatedData models.Reward
		if err := c.ShouldBindJSON(&updatedData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid update data: " + err.Error()})
			return
		}

		existingReward.Name = updatedData.Name
		existingReward.Value = updatedData.Value
		existingReward.Quantity = updatedData.Quantity
		existingReward.Probability = updatedData.Probability

		if err := db.Save(&existingReward).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update reward"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Reward updated successfully", "reward": existingReward})
	}
}


func DeleteReward(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rewardID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reward ID"})
			return
		}

		if err := db.Delete(&models.Reward{}, uint(rewardID)).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete reward"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Reward deleted successfully"})
	}
}