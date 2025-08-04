package handlers

import (
    "net/http"
    "gorm.io/gorm"
    "os"
    "time"
    "encoding/json"
    "errors"
    "log"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/kaelCoding/toyBE/internal/models"
    "github.com/kaelCoding/toyBE/internal/utils"
    "github.com/dgrijalva/jwt-go"
)

func RegisterUser(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var newUser models.User
        if err := json.NewDecoder(c.Request.Body).Decode(&newUser); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
            return
        }

        // Hash the password
        hash, err := utils.GenerateFromPassword(newUser.Password, &utils.HashParams{ // Pass newUser.Password and HashParams directly
            Memory:      64 * 1024,
            Iterations:  3,
            Parallelism: 2,
            SaltLength:  16,
            KeyLength:   32,
        })
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        newUser.Password = hash

        // Create the user in the database
        result := db.Create(&newUser)
        if result.Error != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
            return
        }

        c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
    }
}

func LoginUser(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var loginData models.Login
        if err := c.ShouldBindJSON(&loginData); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
            return
        }

        user := &models.User{}
        result := db.Where(&models.User{Username: loginData.Username}).First(user)
        if result.Error != nil {
            if errors.Is(result.Error, gorm.ErrRecordNotFound) {
                c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"}) // Lỗi cơ sở dữ liệu chung
                // Có thể log lỗi chi tiết hơn ở đây để debug
                log.Println("Database error:", result.Error)
            }
            return
        }

        // Sử dụng hàm VerifyPassword từ utils
        match, err := loginData.VerifyPassword(user.Password)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error verifying password"})
            return
        }

        if !match {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
            return
        }

        // Nếu xác thực thành công, tạo token và trả về
        claims := &models.CustomJWTClaims{
            ID:   user.ID,
            Username: user.Username,
            Email: user.Email,
            Admin:    user.Admin,
            StandardClaims: jwt.StandardClaims{
                ExpiresAt: time.Now().Add(time.Minute * 30).Unix(),
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
      var users []models.User
      
      if err := db.Find(&users).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
      }
      
      var userResponses []models.UserResponse
      
      for _, user := range users {
          userResponses = append(userResponses, models.UserResponse{
              ID:       user.ID,
              Username: user.Username,
              Email:    user.Email,
              Admin:    user.Admin,
          })
      }
      
      c.JSON(http.StatusOK, userResponses)
    }
}

func ValidateToken(tokenString string) (*models.CustomJWTClaims, error) {
    jwtSecret := os.Getenv("JWT_SECRET")
    secretKey := []byte(jwtSecret)

    token, err := jwt.ParseWithClaims(tokenString, &models.CustomJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
        return secretKey, nil
    })
    if err != nil {
        if errors.Is(err, jwt.ErrSignatureInvalid) {
        return nil, errors.New("Invalid token signature")
        } else {
        return nil, err // Handle other errors
        }
    }

    claims, ok := token.Claims.(*models.CustomJWTClaims)
    if !ok {
        return nil, errors.New("Invalid token claims")
    }

    // Check if token is expired
    if time.Now().Unix() > claims.ExpiresAt {
        return nil, errors.New("Token expired")
    }

    return claims, nil
}

func GetUserFromDatabase(db *gorm.DB, userID uint) (*models.User, error) {
    var user models.User
    result := db.First(&user, userID)
    if result.Error != nil {
      if errors.Is(result.Error, gorm.ErrRecordNotFound) {
        return nil, errors.New("User not found")
      } else {
        return nil, result.Error // Handle other errors
      }
    }
  
    return &user, nil
}

func GetUser(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
            return
        }

        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
            return
        }

        tokenString := parts[1] // Extract the token

        claims, err := ValidateToken(tokenString) // Now pass the token only
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            return
        }

        user, err := GetUserFromDatabase(db, claims.ID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        userResponse := models.UserResponse{
            ID:        user.ID,
            Username:  user.Username,
            Email:     user.Email,
            Admin:     user.Admin,
        }

        c.JSON(http.StatusOK, userResponse)
    }
}
