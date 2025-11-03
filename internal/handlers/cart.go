package handlers

import (
    "errors"
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/kaelCoding/toyBE/internal/models"
    "gorm.io/gorm"
)

func getOrCreateCart(db *gorm.DB, userID uint) (*models.Cart, error) {
    var cart models.Cart
    if err := db.Where("user_id = ?", userID).First(&cart).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            cart = models.Cart{UserID: userID}
            if err := db.Create(&cart).Error; err != nil {
                return nil, err
            }
            return &cart, nil
        }
        return nil, err
    }
    return &cart, nil
}

func AddToCart(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, exists := c.Get("userID")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
            return
        }

        var req models.AddToCartRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
            return
        }

        if req.Quantity <= 0 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Quantity must be greater than 0"})
            return
        }

        cart, err := getOrCreateCart(db, userID.(uint))
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get or create cart"})
            return
        }

        var existingItem models.CartItem
        if err := db.Where("cart_id = ? AND product_id = ?", cart.ID, req.ProductID).First(&existingItem).Error; err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                newItem := models.CartItem{
                    CartID:    cart.ID,
                    ProductID: req.ProductID,
                    Quantity:  req.Quantity,
                }
                if err := db.Create(&newItem).Error; err != nil {
                    c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item to cart"})
                    return
                }
                c.JSON(http.StatusCreated, newItem)
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking for item"})
            }
        } else {
            existingItem.Quantity += req.Quantity
            if err := db.Save(&existingItem).Error; err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item quantity"})
                return
            }
            c.JSON(http.StatusOK, existingItem)
        }
    }
}

func GetCart(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, exists := c.Get("userID")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
            return
        }

        var cart models.Cart
        err := db.Where("user_id = ?", userID.(uint)).
            Preload("CartItems.Product.Categories").
            First(&cart).Error
            
        if err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                c.JSON(http.StatusOK, models.Cart{UserID: userID.(uint), CartItems: []models.CartItem{}})
                return
            }
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve cart"})
            return
        }

        c.JSON(http.StatusOK, cart)
    }
}

func UpdateCartItemQuantity(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, exists := c.Get("userID")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
            return
        }
        
        cartItemID, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cart item ID"})
            return
        }

        var req models.UpdateCartItemRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
            return
        }

        cart, err := getOrCreateCart(db, userID.(uint))
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cart"})
            return
        }

        var item models.CartItem
        if err := db.Where("id = ? AND cart_id = ?", cartItemID, cart.ID).First(&item).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "Cart item not found or does not belong to user"})
            return
        }

        if req.Quantity <= 0 {
            if err := db.Delete(&item).Error; err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item from cart"})
                return
            }
            c.JSON(http.StatusOK, gin.H{"message": "Item removed from cart"})
        } else {
            item.Quantity = req.Quantity
            if err := db.Save(&item).Error; err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item quantity"})
                return
            }
            c.JSON(http.StatusOK, item)
        }
    }
}

func DeleteCartItem(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, exists := c.Get("userID")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
            return
        }

        cartItemID, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cart item ID"})
            return
        }

        cart, err := getOrCreateCart(db, userID.(uint))
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cart"})
            return
        }

        var item models.CartItem
        if err := db.Where("id = ? AND cart_id = ?", cartItemID, cart.ID).First(&item).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "Cart item not found or does not belong to user"})
            return
        }

        if err := db.Delete(&item).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item from cart"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"message": "Item removed from cart"})
    }
}