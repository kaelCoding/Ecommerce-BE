package main

import (
	"fmt"
	"log"
	"os"
	"github.com/joho/godotenv"
	"github.com/kaelCoding/toyBE/internal/database"
	"github.com/kaelCoding/toyBE/internal/models"
	"github.com/kaelCoding/toyBE/internal/router"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	database.ConnectDB()
	err = database.DB.AutoMigrate(&models.Product{}, &models.Category{})
	if err != nil {
		log.Fatal("Error migrating schema:", err)
	}

	r := router.SetupRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}

	fmt.Printf("Server is running on port %s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(fmt.Sprintf("Error starting server: %w", err))
	}
}