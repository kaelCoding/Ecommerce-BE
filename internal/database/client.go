package database

import (
    "fmt"
    "log"
    "os"
    "strconv"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "github.com/joho/godotenv"
    "github.com/kaelCoding/toyBE/internal/models"
)

var DB *gorm.DB

func GetDB() *gorm.DB {
    return DB
}

func ConnectDB() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    dbHost := os.Getenv("DB_HOST")
    dbUser := os.Getenv("DB_USERNAME")
    dbPass := os.Getenv("DB_PASSWORD")
    dbName := os.Getenv("DB_NAME")
    dbPort := os.Getenv("DB_PORT")

    p, err := strconv.Atoi(dbPort)
    if err != nil {
        panic(err)
    }

    dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=require", dbHost, p, dbUser, dbPass, dbName)
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        panic(fmt.Sprintf("failed to connect database, got error: %v", err))
    }

    err = db.AutoMigrate(&models.Product{}, &models.Category{}, &models.User{})
    if err != nil {
        log.Println("Error migrating schema:", err)
    }

    DB = db
}
