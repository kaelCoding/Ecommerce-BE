package models

import (
    "gorm.io/gorm"
    "gorm.io/datatypes"
)

type Product struct {
    gorm.Model
    ID              uint            `gorm:"primaryKey;autoIncrement" json:"ID"`
    Name            string          `json:"name"`
    Description     string          `gorm:"type:text" json:"description"`
    Price           string          `json:"price"`
    ImageURLs       datatypes.JSON  `json:"image_urls"`
    Categories      []Category      `gorm:"many2many:product_categories;" json:"categories"`
}

type Category struct {
    gorm.Model
    ID          uint        `gorm:"primaryKey;autoIncrement" json:"ID"`
    Name        string      `gorm:"uniqueIndex;size:255" json:"name"`
    Description string      `gorm:"size:255" json:"description"`
    Products    []Product   `gorm:"many2many:product_categories;" json:"products"`
}