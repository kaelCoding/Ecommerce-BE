package handlers

import (
    "log"
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/kaelCoding/toyBE/internal/database"
    "github.com/kaelCoding/toyBE/internal/loyalty"
    "github.com/kaelCoding/toyBE/internal/models"
    "github.com/kaelCoding/toyBE/internal/services"
)

func CreateOrderHandler(c *gin.Context) {
    var req models.OrderRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
        return
    }

    userID, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    db := database.GetDB()
    var user models.User
    if err := db.First(&user, userID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    tx := db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            log.Printf("Recovered from panic: %v", r)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "An internal error occurred"})
        }
    }()
    
    var product models.Product
    if err := tx.First(&product, req.ProductID).Error; err != nil {
        tx.Rollback()
        c.JSON(http.StatusNotFound, gin.H{"error": "Product with ID " + strconv.Itoa(int(req.ProductID)) + " not found"})
        return
    }

    price, _ := strconv.ParseFloat(product.Price, 64)
    originalAmount := price * float64(req.Quantity)

    orderItems := []models.OrderItem{
        {
            ProductID: req.ProductID,
            Quantity:  req.Quantity,
            Price:     price,
        },
    }

    vipInfo := loyalty.GetVIPLevelInfo(user.VIPLevel)
    discountAmount := originalAmount * vipInfo.Discount
    finalAmount := originalAmount - discountAmount

    order := models.Order{
        UserID:          user.ID,
        OriginalAmount:  originalAmount,
        DiscountApplied: discountAmount,
        TotalAmount:     finalAmount, 
        Status:          "completed",
        CustomerName:    req.CustomerName,
        CustomerPhone:   req.CustomerPhone,
        CustomerAddress: req.CustomerAddress,
        CustomerEmail:   req.CustomerEmail,
        PaymentMethod:   req.PaymentMethod,
        OrderItems:      orderItems,
    }

    if err := tx.Create(&order).Error; err != nil {
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
        return
    }

    if err := loyalty.UpdateUserLoyaltyStatus(tx, user.ID, order.TotalAmount); err != nil {
        tx.Rollback()
        log.Printf("Failed to update loyalty status for user %d: %v", user.ID, err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update loyalty status"})
        return
    }

    if err := tx.Commit().Error; err != nil {
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
        return
    }

    go func() {
        var fullOrder models.Order
        if err := db.Preload("User").Preload("OrderItems.Product").First(&fullOrder, order.ID).Error; err != nil {
            log.Printf("Error fetching full order details for email (OrderID: %d): %v", order.ID, err)
            return
        }

        if err := services.SendOrderConfirmationEmail(fullOrder); err != nil {
            log.Printf("Failed to send order confirmation email to admin (OrderID: %d): %v", fullOrder.ID, err)
        }

        if fullOrder.CustomerEmail != "" {
            if err := services.SendInvoiceToCustomer(fullOrder, fullOrder.CustomerEmail); err != nil {
                log.Printf("Failed to send invoice email to customer (OrderID: %d): %v", fullOrder.ID, err)
            }
        }
    }()

    c.JSON(http.StatusCreated, gin.H{
        "message": "Order created successfully. Confirmation emails are being sent.",
        "order":   order,
    })
}
