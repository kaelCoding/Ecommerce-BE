package models

import (
    "gorm.io/gorm"
    "time"
)

type Message struct {
    gorm.Model
    SenderID   uint   `json:"senderId"`
    Sender     User   `gorm:"foreignKey:SenderID"`
    ReceiverID uint   `json:"receiverId"`
    Receiver   User   `gorm:"foreignKey:ReceiverID"`
    Content    string `json:"content"`
    Read       bool   `json:"read" gorm:"default:false"`
    Timestamp  time.Time `json:"timestamp"`
}