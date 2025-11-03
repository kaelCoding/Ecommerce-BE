package models

import (
    "gorm.io/datatypes"
    "gorm.io/gorm"
)

type ProxyOrder struct {
    gorm.Model
    UserID          uint   `json:"userId"` 
    User            User   `json:"user"`
    MercariURL      string `gorm:"type:text" json:"mercariURL"`      
    MercariItemID   string `gorm:"index" json:"mercariItemID"`  
    ProductName     string `json:"productName"`
    ProductPriceJPY float64 `json:"productPriceJPY"`  
    ProductCondition string `json:"productCondition"`
    ProductDescription string `gorm:"type:text" json:"productDescription"`
    ImageURLs       datatypes.JSON `json:"imageURLs"`
    CustomerName    string `json:"customerName"`
    CustomerPhone   string `json:"customerPhone"`
    CustomerAddress string `json:"customerAddress"`
    CustomerEmail   string `json:"customerEmail"`
    ExchangeRate  float64 `json:"exchangeRate"`  
    ServiceFee    float64 `json:"serviceFee"`   
    TotalAmountVND float64 `json:"totalAmountVND"` 
    Status        string  `gorm:"default:'pending_quote'" json:"status"`
    Quantity      int    `json:"quantity"`
    PaymentMethod string `json:"paymentMethod"`
}