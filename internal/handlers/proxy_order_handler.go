package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kaelCoding/toyBE/internal/models"
	"github.com/kaelCoding/toyBE/internal/services"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Khai báo hằng số tỷ giá và phí dịch vụ
const (
	mercariExchangeRate     = 185.0 // Tỷ giá mới: 1 JPY = 185 VND
	mercariServiceFeePercent = 0.05  // Phí dịch vụ 5%
)

type CreateProxyOrderRequest struct {
	MercariURL         string   `json:"mercariURL" binding:"required"`
	MercariItemID      string   `json:"mercariItemID" binding:"required"`
	ProductName        string   `json:"productName" binding:"required"`
	ProductPriceJPY    float64  `json:"productPriceJPY"`
	ProductCondition   string   `json:"productCondition"`
	ProductDescription string   `json:"productDescription"`
	ImageURLs          []string `json:"imageURLs"`
	CustomerName       string   `json:"customerName" binding:"required"`
	CustomerPhone      string   `json:"customerPhone" binding:"required"`
	CustomerAddress    string   `json:"customerAddress" binding:"required"`
	CustomerEmail      string   `json:"customerEmail" binding:"required"`
	Quantity           int      `json:"quantity" binding:"required"`
	PaymentMethod      string   `json:"paymentMethod" binding:"required"`
	// Đã loại bỏ các trường giá VNĐ do backend sẽ tự tính
	// ExchangeRate       float64  `json:"exchangeRate"`
	// ServiceFee         float64  `json:"serviceFee"`
	// TotalAmountVND     float64  `json:"totalAmountVND"`
}

func CreateProxyOrder(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateProxyOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
			return
		}

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

		imageURLsJSON, err := json.Marshal(req.ImageURLs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process image URLs"})
			return
		}

		// --- LOGIC TÍNH GIÁ MỚI TẠI BACKEND ---
		// 1. Tính giá cơ bản (quy đổi)
		basePriceVND := req.ProductPriceJPY * mercariExchangeRate
		// 2. Tính phí dịch vụ (5% của giá cơ bản)
		serviceFeeVND := basePriceVND * mercariServiceFeePercent
		// 3. Tính tổng tiền cho 1 sản phẩm (chưa ship)
		singleItemTotalVND := basePriceVND + serviceFeeVND
		// 4. Tính tổng tiền cuối cùng (theo số lượng)
		finalTotalAmountVND := singleItemTotalVND * float64(req.Quantity)
		// --- KẾT THÚC LOGIC TÍNH GIÁ ---

		proxyOrder := models.ProxyOrder{
			UserID:             user.ID,
			MercariURL:         req.MercariURL,
			MercariItemID:      req.MercariItemID,
			ProductName:        req.ProductName,
			ProductPriceJPY:    req.ProductPriceJPY,
			ProductCondition:   req.ProductCondition,
			ProductDescription: req.ProductDescription,
			ImageURLs:          datatypes.JSON(imageURLsJSON),
			CustomerName:       req.CustomerName,
			CustomerPhone:      req.CustomerPhone,
			CustomerAddress:    req.CustomerAddress,
			CustomerEmail:      req.CustomerEmail,
			// Lưu trữ các giá trị đã tính toán
			ExchangeRate:       mercariExchangeRate,
			ServiceFee:         serviceFeeVND, // Phí dịch vụ cho 1 sản phẩm
			TotalAmountVND:     finalTotalAmountVND, // Tổng tiền cuối cùng
			Status:             "pending_quote",
			Quantity:           req.Quantity,
			PaymentMethod:      req.PaymentMethod,
		}

		if err := db.Create(&proxyOrder).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create proxy order"})
			return
		}

		go func() {
			db.Preload("User").First(&proxyOrder, proxyOrder.ID)
			
			// Gọi hàm email mới
			if err := services.SendProxyOrderConfirmationEmail(proxyOrder); err != nil {
				log.Printf("Failed to send proxy order confirmation email to admin (OrderID: %d): %v", proxyOrder.ID, err)
			}
			if proxyOrder.CustomerEmail != "" {
				// Gọi hàm email mới
				if err := services.SendProxyInvoiceToCustomer(proxyOrder, proxyOrder.CustomerEmail); err != nil {
					log.Printf("Failed to send proxy invoice email to customer (OrderID: %d): %v", proxyOrder.ID, err)
				}
			}
		}()

		c.JSON(http.StatusCreated, gin.H{
			"message": "Proxy order created successfully. Confirmation emails are being sent.",
			"order":   proxyOrder,
		})
	}
}