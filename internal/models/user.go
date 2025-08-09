package models

import (
	"gorm.io/gorm"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kaelCoding/toyBE/internal/utils"
)

type User struct {
	gorm.Model
	Username   string `gorm:"unique;not null;index" json:"username"`
	Email      string `gorm:"unique;not null;index" json:"email"`
	Password   string `gorm:"not null" json:"-"` 
	Admin      bool   `gorm:"default:false" json:"admin"`
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
