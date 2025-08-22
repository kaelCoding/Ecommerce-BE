package models

import (
	"time"
	"gorm.io/gorm"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kaelCoding/toyBE/internal/utils"
)

type User struct {
	gorm.Model
	Username   			string `gorm:"unique;not null;index" json:"username"`
	Email      			string `gorm:"unique;not null;index" json:"email"`
	Password   			string `gorm:"not null" json:"password"` 
	Admin      			bool   `gorm:"default:false" json:"admin"`
	FCMToken   			string `json:"fcmToken"`
	TotalSpent          float64    `gorm:"default:0" json:"totalSpent"`
	VIPLevel            int        `gorm:"default:0" json:"vipLevel"`
	VIPExpiryDate       *time.Time `json:"vipExpiryDate"` 
	MaintenanceSpending float64    `gorm:"default:0" json:"maintenanceSpending"`
	DiscountPercentage  float64    `gorm:"default:0" json:"discountPercentage"`
}

type UserProfileResponse struct {
	ID                  uint       `json:"id"`
	Username            string     `json:"username"`
	Email               string     `json:"email"`
	Admin               bool       `json:"admin"`
	TotalSpent          float64    `json:"totalSpent"`
	VIPLevel            int        `json:"vipLevel"`
	VIPExpiryDate       *time.Time `json:"vipExpiryDate"`
	DiscountPercentage  float64    `json:"discountPercentage"`
	NextLevelRequirement float64   `json:"nextLevelRequirement"`
	MaintenanceRequirement float64 `json:"maintenanceRequirement"`
}

type UserResponse struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Admin     bool   `json:"admin"`
}

type Login struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type CustomJWTClaims struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Admin    bool   `json:"admin"`
	jwt.RegisteredClaims
}

func (u *User) VerifyPassword(password string) (bool, error) {
	return utils.ComparePasswordAndHash(password, u.Password)
}
