package handlers

import (
	"log"
	"net/http"
	"strconv"
	"gorm.io/gorm"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kaelCoding/toyBE/internal/database"
	"github.com/kaelCoding/toyBE/internal/loyalty"
	"github.com/kaelCoding/toyBE/internal/models"
	"github.com/kaelCoding/toyBE/internal/services"
)

func generateShippingCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 10)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func GetAllOrders(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var orders []models.Order
		if err := db.Preload("User").Order("created_at desc").Find(&orders).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve orders"})
			return
		}
		c.JSON(http.StatusOK, orders)
	}
}

func CreateOrderFromCart(c *gin.Context) {
	var req models.CartCheckoutRequest
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
	
	var cart models.Cart
	if err := db.Where("user_id = ?", userID).Preload("CartItems.Product").First(&cart).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
		return
	}

	if len(cart.CartItems) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cart is empty"})
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

	var orderItems []models.OrderItem
	var originalAmount float64 = 0

	for _, item := range cart.CartItems {
		price, _ := strconv.ParseFloat(item.Product.Price, 64)
		itemTotal := price * float64(item.Quantity)
		originalAmount += itemTotal
		
		orderItems = append(orderItems, models.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     price,
		})
	}

	vipInfo := loyalty.GetVIPLevelInfo(user.VIPLevel)
	discountAmount := originalAmount * vipInfo.Discount
	finalAmount := originalAmount - discountAmount

	shippingCode := generateShippingCode()

	order := models.Order{
		UserID:          user.ID,
		OriginalAmount:  originalAmount,
		DiscountApplied: discountAmount,
		TotalAmount:     finalAmount,
		Status:          "completed", 
		CustomerName:    req.CustomerName, 
		CustomerPhone:   req.CustomerPhone,
		CustomerAddress: req.CustomerAddress,
		CustomerEmail:   user.Email, 
		PaymentMethod:   req.PaymentMethod,
		OrderItems:      orderItems,
		ShippingCode:    shippingCode,
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order: " + err.Error()})
		return
	}

	if err := loyalty.UpdateUserLoyaltyStatus(tx, user.ID, order.TotalAmount); err != nil {
		tx.Rollback()
		log.Printf("Failed to update loyalty status for user %d: %v", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update loyalty status"})
		return
	}

	if err := tx.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear cart"})
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
		"message": "Order created successfully from cart. Confirmation emails are being sent.",
		"order":   order,
	})
}