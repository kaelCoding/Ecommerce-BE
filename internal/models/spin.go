package models

import (
	"time"
	"gorm.io/gorm"
)

type Reward struct {
	gorm.Model
	Name        string  `gorm:"not null" json:"name"`
	Value       string  `json:"value"`
	Quantity    int     `gorm:"default:0" json:"quantity"`
	Probability float64 `gorm:"not null" json:"probability"`
}

type SpinLog struct {
	gorm.Model
	OrderID  uint      `gorm:"not null" json:"orderId"`
	UserID   uint      `gorm:"not null" json:"userId"`
	RewardID uint      `gorm:"not null" json:"rewardId"`
	SpinDate time.Time `gorm:"not null" json:"spinDate"`

	Order  Order  `gorm:"foreignKey:OrderID" json:"order"`
	User   User   `gorm:"foreignKey:UserID" json:"user"`
	Reward Reward `gorm:"foreignKey:RewardID" json:"reward"`
}

type SpinRequest struct {
	ShippingCode string `json:"shippingCode" binding:"required"`
}

type SpinResponse struct {
	Message string `json:"message"`
	Reward  Reward `json:"reward"`
}