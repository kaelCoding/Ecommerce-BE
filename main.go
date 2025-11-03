package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kaelCoding/toyBE/internal/database"
	"github.com/kaelCoding/toyBE/internal/models"
	"github.com/kaelCoding/toyBE/internal/router"
	"github.com/kaelCoding/toyBE/internal/pkg/r2"
	"github.com/kaelCoding/toyBE/internal/chat"
	"github.com/kaelCoding/toyBE/internal/loyalty"
    "github.com/robfig/cron/v3"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Cảnh báo: Không tìm thấy file .env, sẽ sử dụng biến môi trường hệ thống. Lỗi: %v\n", err)
	}

	r2.Init()
	database.ConnectDB()
	db := database.GetDB()

	fmt.Println("Migrating database schemas...")

	err = database.DB.AutoMigrate(&models.User{}, &models.Product{}, &models.Category{}, &models.Message{}, &models.Order{}, &models.OrderItem{}, &models.Reward{}, &models.SpinLog{}, &models.ProxyOrder{}, &models.Cart{}, &models.CartItem{})
	if err != nil {
		log.Fatal("Error migrating schema:", err)
	}
	fmt.Println("Database migration successful.")

	// database.CreateInitialAdmin(database.DB)

	c := cron.New()
	c.AddFunc("0 1 * * *", func() { loyalty.CheckAndApplyDemotions(db) })
	c.Start()
	log.Println("Cron job for VIP demotion checks scheduled.")

	hub := chat.NewHub()
	go hub.Run()

	r := router.SetupRouter(hub)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server is running on http://localhost:%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
