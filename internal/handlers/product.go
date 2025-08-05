package handlers

import (
    "net/http"
    "strconv"
    "errors"
    "gorm.io/gorm"
    "fmt"
    "time"
    "encoding/json"
	"path/filepath"

    
    "github.com/gin-gonic/gin"
	"github.com/kaelCoding/toyBE/internal/models"
	"github.com/kaelCoding/toyBE/internal/database"
)

func AddProduct(c *gin.Context) {
    db := database.GetDB()

    err := c.Request.ParseMultipartForm(10 << 20) // 10 * 1024 * 1024
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing form: " + err.Error()})
        return
    }

    name := c.PostForm("name")
    description := c.PostForm("description")
    price := c.PostForm("price")
    categoryID := c.PostForm("category_id")

    if name == "" || price == "" || categoryID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Name, price, and category_id are required fields"})
        return
    }

    form, err := c.MultipartForm()
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Error getting multipart form: " + err.Error()})
        return
    }
    
    files := form.File["images"] 
    imageURLs := []string{}

    for _, file := range files {
        uniqueFileName := fmt.Sprintf("%d-%s", time.Now().UnixNano(), filepath.Base(file.Filename))
        dst := filepath.Join("./uploads", uniqueFileName)

        if err := c.SaveUploadedFile(file, dst); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to save file: " + err.Error()})
            return
        }

        // fileURL := fmt.Sprintf("https://ecommerce-be-w27g.onrender.com/uploads/%s", uniqueFileName)
        fileURL := fmt.Sprintf("https://ecommerce-be-w27g.onrender.com/uploads/%s", uniqueFileName)
        
        imageURLs = append(imageURLs, fileURL)
    }

    imageURLsJSON, err := json.Marshal(imageURLs)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating image URLs JSON"})
        return
    }
    
    var catID uint
    fmt.Sscanf(categoryID, "%d", &catID)
    
    product := models.Product{
        Name:        name,
        Description: description,
        Price:       price,
        CategoryID:  catID,
        ImageURLs:   imageURLsJSON,
    }

    // 5. Kiểm tra Category tồn tại
    var category models.Category
    if err := db.First(&category, product.CategoryID).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // 6. Lưu sản phẩm vào DB
    if err := db.Create(&product).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product: " + err.Error()})
        return
    }
    
    // Trả về dữ liệu sản phẩm đã tạo thành công
    c.JSON(http.StatusCreated, gin.H{"data": product})
}


func GetProducts(c *gin.Context) {
    db := database.GetDB()
    var products []models.Product

    if err := db.Preload("Category").Order("created_at DESC").Find(&products).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    for i := range products {
        if products[i].Category.ID != 0 {
            products[i].CategoryName = products[i].Category.Name
        }
    }

    c.JSON(http.StatusOK, products)
}

func GetProductByID(c *gin.Context) {
    idStr := c.Param("id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
        return
    }

    db := database.GetDB()
    var product models.Product

    if err := db.Preload("Category").First(&product, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    if product.Category.ID != 0 {
        product.CategoryName = product.Category.Name
    }

    c.JSON(http.StatusOK, product)
}

func UpdateProduct(c *gin.Context) {
    db := database.GetDB() // Lấy instance DB

    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
        return
    }

    var existingProduct models.Product
    if err := db.First(&existingProduct, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }

    err = c.Request.ParseMultipartForm(10 << 20) // Giới hạn 10MB
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing form data"})
        return
    }

    existingProduct.Name = c.PostForm("name")
    existingProduct.Description = c.PostForm("description")
    existingProduct.Price = c.PostForm("price")

    catID, _ := strconv.Atoi(c.PostForm("category_id"))
    existingProduct.CategoryID = uint(catID)

    form, err := c.MultipartForm()
    if err == nil {
        files := form.File["images"]
        if len(files) > 0 {
            var newImageURLs []string
            
            for _, file := range files {
                uniqueFileName := fmt.Sprintf("%d-%s", time.Now().UnixNano(), filepath.Base(file.Filename))
                dst := filepath.Join("./uploads", uniqueFileName)

                if err := c.SaveUploadedFile(file, dst); err != nil {
                    c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to save file"})
                    return
                }
                fileURL := fmt.Sprintf("https://ecommerce-be-w27g.onrender.com/uploads/%s", uniqueFileName)
                newImageURLs = append(newImageURLs, fileURL)
            }
            
            imageURLsJSON, _ := json.Marshal(newImageURLs)
            existingProduct.ImageURLs = imageURLsJSON
        }
    }

    if err := db.Save(&existingProduct).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": existingProduct})
}

func DeleteProduct(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
        return
    }

    var product models.Product
    if err := database.DB.Delete(&product, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Product deleted"})
}