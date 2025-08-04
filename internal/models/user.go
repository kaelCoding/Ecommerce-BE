package models

import (
    "gorm.io/gorm"
    "github.com/dgrijalva/jwt-go"
    "github.com/kaelCoding/toyBE/internal/utils"
)

type User struct {
    gorm.Model
    ID        uint  `gorm:"primaryKey;autoIncrement" json:"id"`
    Username string `gorm:"unique;not null" json:"username"`
    Email    string `gorm:"unique;not null" json:"email"`
    Password string `gorm:"not null" json:"password"`
    Admin    bool   `json:"admin"` 
}

type UserResponse struct {
    gorm.Model
    ID        uint      `json:"id"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    Admin    bool   `json:"admin"` 
}

type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CustomJWTClaims struct {
    ID       uint   `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Admin    bool   `json:"admin"`
	jwt.StandardClaims
}

func (l Login) VerifyPassword(hashPassword string) (match bool, err error) {
	match, err = utils.ComparePasswordAndHash(l.Password, hashPassword)
	if err != nil {
		return false, err
	}
	return match, err
}