package database

import (
	"fmt"
	"log"

	"github.com/kaelCoding/toyBE/internal/models"
	"github.com/kaelCoding/toyBE/internal/utils"
	"gorm.io/gorm"
)

func CreateInitialAdmin(db *gorm.DB) {
	var admin models.User
	result := db.Where("admin = ?", true).First(&admin)

	if result.Error == gorm.ErrRecordNotFound {
		fmt.Println("No admin user found. Creating initial admin user...")

		hash, err := utils.GenerateFromPassword("admin_password_123", &utils.HashParams{
			Memory:      64 * 1024,
			Iterations:  4,
			Parallelism: 2,
			SaltLength:  16,
			KeyLength:   32,
		})
		if err != nil {
			log.Fatalf("Failed to hash initial admin password: %v", err)
		}

		newAdmin := models.User{
			Username: "Admin",
			Email:    "kael@gmail.com",
			Password: hash,
			Admin:    true,
		}

		if err := db.Create(&newAdmin).Error; err != nil {
			log.Fatalf("Failed to create initial admin user: %v", err)
		}

		fmt.Println("Initial admin user created successfully.")
	} else if result.Error != nil {
		log.Fatalf("Database error when checking for admin user: %v", result.Error)
	} else {
		fmt.Println("Admin user already exists. Skipping creation.")
	}
}