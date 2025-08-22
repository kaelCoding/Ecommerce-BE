package handlers

import (
    "errors"
    "log"
    "net/http"
    "os"
    "strings"
    "time"
    "fmt"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "github.com/kaelCoding/toyBE/internal/models"
    "github.com/kaelCoding/toyBE/internal/utils"
    "github.com/kaelCoding/toyBE/internal/loyalty"
    "gorm.io/gorm"
)

func RegisterUser(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var newUser models.User

        if err := c.ShouldBindJSON(&newUser); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
            return
        }

        hash, err := utils.GenerateFromPassword(newUser.Password, &utils.HashParams{
            Memory:      64 * 1024,
            Iterations:  4,
            Parallelism: 2,
            SaltLength:  16,
            KeyLength:   32,
        })
        if err != nil {
            log.Printf("Error hashing password: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
            return
        }
        newUser.Password = hash

        result := db.Create(&newUser)
        if result.Error != nil {
            if strings.Contains(result.Error.Error(), "UNIQUE constraint failed") {
                c.JSON(http.StatusConflict, gin.H{"error": "Username or email already exists"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
            }
            return
        }

        userResponse := models.UserResponse{
            ID:       newUser.ID,
            Username: newUser.Username,
            Email:    newUser.Email,
            Admin:    newUser.Admin,
        }

        c.JSON(http.StatusCreated, gin.H{
            "message": "User created successfully",
            "user":    userResponse,
        })
    }
}

func LoginUser(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var loginData models.Login
        if err := c.ShouldBindJSON(&loginData); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
            return
        }

        user := &models.User{}
        result := db.Where("email = ?", loginData.Email).First(user)
        if result.Error != nil {
            if errors.Is(result.Error, gorm.ErrRecordNotFound) {
                c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
                log.Println("Database error:", result.Error)
            }
            return
        }

        match, err := user.VerifyPassword(loginData.Password)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error verifying password"})
            return
        }

        if !match {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
            return
        }

        claims := &models.CustomJWTClaims{
            ID:       user.ID,
            Username: user.Username,
            Email:    user.Email,
            Admin:    user.Admin,
            RegisteredClaims: jwt.RegisteredClaims{
                ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
                IssuedAt:  jwt.NewNumericDate(time.Now()),
                NotBefore: jwt.NewNumericDate(time.Now()),
            },
        }

        token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
        jwtSecret := os.Getenv("JWT_SECRET")
        tokenString, err := token.SignedString([]byte(jwtSecret))
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
            return
        }

        c.JSON(http.StatusOK, gin.H{"token": tokenString})
    }
}

func GetAllUsers(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var userResponses []models.UserResponse

        if err := db.Model(&models.User{}).Select("id, username, email, admin").Find(&userResponses).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, userResponses)
    }
}

func GetUser(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, exists := c.Get("userID")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
            return
        }

        var user models.User
        if err := db.First(&user, userID).Error; err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
            }
            return
        }
        
        if user.VIPLevel > 0 && user.VIPLevel < 4 && user.VIPExpiryDate != nil && time.Now().After(*user.VIPExpiryDate) {
            log.Printf("Lazy demotion check for User ID %d", user.ID)
            user.VIPLevel--
            user.MaintenanceSpending = 0
            if user.VIPLevel > 0 {
                newExpiryDate := time.Now().AddDate(0, 3, 0)
                user.VIPExpiryDate = &newExpiryDate
            } else {
                user.VIPExpiryDate = nil
            }
            db.Save(&user)
        }
        
        currentVIPInfo := loyalty.GetVIPLevelInfo(user.VIPLevel)
        nextLevelInfo := loyalty.GetVIPLevelInfo(user.VIPLevel + 1)

        var nextLevelRequirement float64
        if user.VIPLevel < 4 {
            nextLevelRequirement = nextLevelInfo.Threshold - user.TotalSpent
        }
        
        userProfile := models.UserProfileResponse{
            ID:                     user.ID,
            Username:               user.Username,
            Email:                  user.Email,
            Admin:                  user.Admin,
            TotalSpent:             user.TotalSpent,
            VIPLevel:               user.VIPLevel,
            VIPExpiryDate:          user.VIPExpiryDate,
            DiscountPercentage:     currentVIPInfo.Discount,
            NextLevelRequirement:   nextLevelRequirement,
            MaintenanceRequirement: currentVIPInfo.MaintenanceRequirement - user.MaintenanceSpending,
        }

        c.JSON(http.StatusOK, userProfile)
    }
}

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        var tokenString string

        if c.Query("token") != "" {
            tokenString = c.Query("token")
        } else {
            authHeader := c.GetHeader("Authorization")
            if authHeader == "" {
                c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
                return
            }

            parts := strings.Split(authHeader, " ")
            if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
                c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
                return
            }
            tokenString = parts[1]
        }

        if tokenString == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication token not provided"})
            return
        }

        jwtSecret := os.Getenv("JWT_SECRET")
        claims := &models.CustomJWTClaims{}
        token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
            }
            return []byte(jwtSecret), nil
        })

        if err != nil || !token.Valid {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
            return
        }

        c.Set("userID", claims.ID)
        c.Set("username", claims.Username)
        c.Set("email", claims.Email)
        c.Set("isAdmin", claims.Admin)

        c.Next()
    }
}

func AdminOnlyMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        isAdmin, exists := c.Get("isAdmin")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "User admin status not found in context"})
            c.Abort()
            return
        }

        if !isAdmin.(bool) {
            c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: requires admin privileges"})
            c.Abort()
            return
        }

        c.Next()
    }
}

func GetAdminInfo(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var adminUser models.User
        if err := db.Where("admin = ?", true).First(&adminUser).Error; err != nil {
            if err == gorm.ErrRecordNotFound {
                c.JSON(http.StatusNotFound, gin.H{"error": "Admin user not found"})
                return
            }
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error while finding admin"})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "id":       adminUser.ID,
            "username": adminUser.Username,
        })
    }
}
