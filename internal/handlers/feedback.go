package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/kaelCoding/toyBE/internal/models"
    "github.com/kaelCoding/toyBE/internal/services" 
)

func SendFeedbackHandler(c *gin.Context) {
    var feedback models.Feedback

    if err := c.ShouldBindJSON(&feedback); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feedback data: " + err.Error()})
        return
    }

    if err := services.SendFeedbackEmail(feedback); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send feedback email: " + err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Feedback received successfully and email sent."})
}