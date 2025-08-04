package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kaelCoding/toyBE/internal/models" 
	"github.com/kaelCoding/toyBE/internal/services" 
)

func CreateOrderHandler(c *gin.Context) {
	var order models.Order

	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}

	if err := services.SendOrderConfirmationEmail(order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send order email: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Order received successfully and email sent."})
}