// handlers/order.go
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kaelCoding/toyBE/internal/models" // Đường dẫn tới models
	"github.com/kaelCoding/toyBE/internal/services" // Đường dẫn tới services
)

func CreateOrderHandler(c *gin.Context) {
	var order models.Order

	// Bind JSON từ request vào struct Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}

	// Gọi service để gửi email
	if err := services.SendOrderConfirmationEmail(order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send order email: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Order received successfully and email sent."})
}