package models

import (
	"gorm.io/gorm"
)

type OrderRequest struct {
	ProductID       uint   `json:"productId" binding:"required"`
	Quantity        int    `json:"quantity" binding:"required"`
	CustomerName    string `json:"customerName" binding:"required"`
	CustomerPhone   string `json:"customerPhone" binding:"required"`
	CustomerAddress string `json:"customerAddress" binding:"required"`
	CustomerEmail   string `json:"customerEmail" binding:"required"`
	PaymentMethod   string `json:"paymentMethod" binding:"required"`
}

type Order struct {
	gorm.Model
	UserID          uint        `json:"userId"`
	User            User        `json:"user"`
	TotalAmount     float64     `json:"totalAmount"`
	OriginalAmount  float64     `json:"originalAmount"`
	DiscountApplied float64     `json:"discountApplied"`
	Status          string      `gorm:"default:'pending'" json:"status"`
	CustomerName    string      `json:"customerName"`
	CustomerPhone   string      `json:"customerPhone"`
	CustomerAddress string      `json:"customerAddress"`
	CustomerEmail   string      `json:"customerEmail"`
	PaymentMethod   string      `json:"paymentMethod"`
	OrderItems      []OrderItem `gorm:"foreignKey:OrderID" json:"orderItems"`
	ShippingCode 	string 		`gorm:"unique;index" json:"shippingCode"`
	HasSpun      	bool   		`gorm:"default:false" json:"hasSpun"`
}

type OrderItem struct {
	gorm.Model
	OrderID   uint    `json:"orderId"`
	ProductID uint    `json:"productId"`
	Product   Product `json:"product"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}