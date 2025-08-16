package models

type Order struct {
	ProductName     string `json:"productName" binding:"required"`
	Quantity        int    `json:"quantity" binding:"required"`
	TotalPrice      int    `json:"totalPrice" binding:"required"`
	CustomerName    string `json:"customerName" binding:"required"`
	CustomerPhone   string `json:"customerPhone" binding:"required"`
	CustomerAddress string `json:"customerAddress" binding:"required"`
	PaymentMethod   string `json:"paymentMethod" binding:"required"`
	CustomerEmail   string `json:"customerEmail"`
}