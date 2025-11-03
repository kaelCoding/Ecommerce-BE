package models

import (
	"gorm.io/gorm"
)

type Cart struct {
	gorm.Model
	UserID    uint       `gorm:"uniqueIndex;not null" json:"userId"` 
	User      User       `gorm:"foreignKey:UserID" json:"user"`
	CartItems []CartItem `gorm:"foreignKey:CartID" json:"cartItems"`
}

type CartItem struct {
	gorm.Model
	CartID    uint    `gorm:"index" json:"cartId"` 
	ProductID uint    `json:"productId"`           
	Product   Product `gorm:"foreignKey:ProductID" json:"product"`
	Quantity  int     `json:"quantity"`
}

type AddToCartRequest struct {
	ProductID uint `json:"productId" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity"` 
}

type CartCheckoutRequest struct {
	CustomerName    string `json:"customerName" binding:"required"` 
	CustomerPhone   string `json:"customerPhone" binding:"required"`
	CustomerAddress string `json:"customerAddress" binding:"required"`
	PaymentMethod   string `json:"paymentMethod" binding:"required"`
}
