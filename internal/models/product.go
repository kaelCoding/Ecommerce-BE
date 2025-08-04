package models

import (
    "gorm.io/gorm"
    "gorm.io/datatypes"
)

type Product struct {
    gorm.Model
    ID              uint            `gorm:"primaryKey;autoIncrement" json:"ID"`
    Name            string          `json:"name"`
    Description     string          `gorm:"size:255" json:"description"`
    Price           string          `json:"price"`
    CategoryID      uint            `gorm:"foreignKey:CategoryID" json:"category_id"`
    ImageURLs       datatypes.JSON  `json:"image_urls"`

    Category        Category        `gorm:"foreignKey:CategoryID" json:"category"`
    CategoryName    string          `gorm:"-" json:"category_name"`
}

type Category struct {
    gorm.Model
    ID          uint        `gorm.Model:"primaryKey;autoIncrement" json:"ID"`
    Name        string      `gorm:"uniqueIndex;size:255" json:"name"`
    Description string      `gorm:"size:255" json:"description"`
    Products    []Product   `gorm:"foreignKey:CategoryID;references:ID" json:"product"`
}

