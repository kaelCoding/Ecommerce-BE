package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kaelCoding/toyBE/internal/database"
	"github.com/kaelCoding/toyBE/internal/models"
	"github.com/kaelCoding/toyBE/internal/router"
	"github.com/kaelCoding/toyBE/internal/pkg/cloudinary"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cloudinary.Init()

	database.ConnectDB()

	fmt.Println("Migrating database schemas...")
	err = database.DB.AutoMigrate(&models.User{}, &models.Product{}, &models.Category{})
	if err != nil {
		log.Fatal("Error migrating schema:", err)
	}
	fmt.Println("Database migration successful.")

	database.CreateInitialAdmin(database.DB)

	r := router.SetupRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server is running on http://localhost:%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
