package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kaelCoding/toyBE/internal/database"
	"github.com/kaelCoding/toyBE/internal/models"
	"github.com/kaelCoding/toyBE/internal/pkg/r2"
	"gorm.io/gorm"
)

// THAY ĐỔI: Helper này giờ preload "Categories" (số nhiều)
func createProductResponse(db *gorm.DB, product *models.Product) (map[string]interface{}, error) {
    // Tải lại sản phẩm với danh sách Categories
    if err := db.Preload("Categories").First(&product, product.ID).Error; err != nil {
        return nil, err
    }

    var imageURLs []string
    if product.ImageURLs != nil {
        json.Unmarshal(product.ImageURLs, &imageURLs)
    }

    response := map[string]interface{}{
        "ID":            product.ID,
        "name":          product.Name,
        "description":   product.Description,
        "price":         product.Price,
        "categories":    product.Categories, // Trả về mảng categories
        "image_urls":    imageURLs,         
        "CreatedAt":     product.CreatedAt,
        "UpdatedAt":     product.UpdatedAt,
    }
    return response, nil
}


func AddProduct(c *gin.Context) {
    db := database.GetDB()

    err := c.Request.ParseMultipartForm(10 << 20)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing form: " + err.Error()})
        return
    }

    name := c.PostForm("name")
    description := c.PostForm("description")
    price := c.PostForm("price")
    
    // THAY ĐỔI: Nhận một mảng category IDs
    categoryIDsStr := c.PostFormArray("category_ids")

    if name == "" || price == "" || len(categoryIDsStr) == 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Name, price, and at least one category_id are required fields"})
        return
    }

    // ... (logic xử lý file ảnh giữ nguyên) ...
    form, err := c.MultipartForm()
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Error getting multipart form: " + err.Error()})
        return
    }
    
    files := form.File["images"] 
    imageURLs := []string{}

    for _, file := range files {
        fileContent, err := file.Open()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to open uploaded file"})
            return
        }
        defer fileContent.Close()

        extension := filepath.Ext(file.Filename)
        filename := fmt.Sprintf("product_%d%s", time.Now().UnixNano(), extension)
        contentType := file.Header.Get("Content-Type")

        fileURL, err := r2.UploadToR2(fileContent, "products", filename, contentType)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to upload file to R2: " + err.Error()})
            return
        }
        imageURLs = append(imageURLs, fileURL)
    }

    imageURLsJSON, err := json.Marshal(imageURLs)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating image URLs JSON"})
        return
    }
    
    // THAY ĐỔI: Chuyển đổi mảng string IDs sang mảng uint
    var categoryIDs []uint
    for _, idStr := range categoryIDsStr {
        id, err := strconv.ParseUint(idStr, 10, 32)
        if err == nil {
            categoryIDs = append(categoryIDs, uint(id))
        }
    }

    // Tìm các đối tượng Category
    var categories []models.Category
    if err := db.Find(&categories, categoryIDs).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find categories: " + err.Error()})
        return
    }
    if len(categories) != len(categoryIDs) {
        c.JSON(http.StatusNotFound, gin.H{"error": "One or more categories not found"})
        return
    }
    
    // Tạo sản phẩm (chưa có category)
    product := models.Product{
        Name:        name,
        Description: description,
        Price:       price,
        ImageURLs:   imageURLsJSON,
    }

    if err := db.Create(&product).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product: " + err.Error()})
        return
    }

    // Gán categories cho sản phẩm
    if err := db.Model(&product).Association("Categories").Append(&categories); err != nil {
         c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to associate categories: " + err.Error()})
        return
    }
    
    // Trả về response (đã được cập nhật để preload "Categories")
    response, err := createProductResponse(db, &product)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product details: " + err.Error()})
        return
    }

    c.JSON(http.StatusCreated, response) 
}


func GetProducts(c *gin.Context) {
    db := database.GetDB()
    var products []models.Product

    // THAY ĐỔI: Preload "Categories" (số nhiều)
    if err := db.Preload("Categories").Order("created_at DESC").Find(&products).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // Bỏ vòng lặp gán CategoryName

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

    // THAY ĐỔI: Preload "Categories" (số nhiều)
    if err := db.Preload("Categories").First(&product, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // Bỏ gán CategoryName

    c.JSON(http.StatusOK, product)
}

func GetProductsByCategoryIDWithLimit(c *gin.Context) {
    db := database.GetDB()
    idStr := c.Param("id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
        return
    }

    limitQuery := c.DefaultQuery("limit", "4")
    limit, err := strconv.Atoi(limitQuery)
    if err != nil || limit <= 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter. Must be a positive integer."})
        return
    }

    var products []models.Product
    // THAY ĐỔI: Truy vấn qua bảng trung gian
    err = db.Preload("Categories").
        Joins("JOIN product_categories ON product_categories.product_id = products.id").
        Where("product_categories.category_id = ?", id).
        Order("products.created_at DESC").
        Limit(limit).
        Find(&products).Error

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve products"})
        return
    }

    c.JSON(http.StatusOK, products)
}

func UpdateProduct(c *gin.Context) {
    db := database.GetDB()

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

    err = c.Request.ParseMultipartForm(10 << 20) 
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing form data"})
        return
    }

    existingProduct.Name = c.PostForm("name")
    existingProduct.Description = c.PostForm("description")
    existingProduct.Price = c.PostForm("price")

    // THAY ĐỔI: Nhận mảng category IDs
    categoryIDsStr := c.PostFormArray("category_ids")
    var categoryIDs []uint
    for _, idStr := range categoryIDsStr {
        id, err := strconv.ParseUint(idStr, 10, 32)
        if err == nil {
            categoryIDs = append(categoryIDs, uint(id))
        }
    }

    // Tìm các đối tượng Category mới
    var categories []models.Category
    if len(categoryIDs) > 0 {
        if err := db.Find(&categories, categoryIDs).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find categories: " + err.Error()})
            return
        }
    }
    
    // ... (logic xử lý file ảnh giữ nguyên) ...
    form, err := c.MultipartForm()
    if err == nil {
        files := form.File["images"]
        if len(files) > 0 {
            var newImageURLs []string
            
            for _, file := range files {
                fileContent, err := file.Open()
                if err != nil {
                    return
                }
                defer fileContent.Close()

                extension := filepath.Ext(file.Filename)
                filename := fmt.Sprintf("product_update_%d%s", time.Now().UnixNano(), extension)
                contentType := file.Header.Get("Content-Type")
                
                fileURL, err := r2.UploadToR2(fileContent, "products", filename, contentType)
                if err != nil {
                    c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to upload file to R2"})
                    return
                }
                newImageURLs = append(newImageURLs, fileURL)
            }
            
            imageURLsJSON, _ := json.Marshal(newImageURLs)
            existingProduct.ImageURLs = imageURLsJSON
        }
    }

    // Lưu các trường product cơ bản
    if err := db.Save(&existingProduct).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
        return
    }

    // THAY ĐỔI: Cập nhật (thay thế) các categories liên quan
    if err := db.Model(&existingProduct).Association("Categories").Replace(&categories); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update categories association: " + err.Error()})
        return
    }

    response, err := createProductResponse(db, &existingProduct)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product details: " + err.Error()})
        return
    }

    c.JSON(http.StatusOK, response)
}

func DeleteProduct(c *gin.Context) {
    db := database.GetDB() // Lấy DB instance
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
        return
    }

    var product models.Product
    // THAY ĐỔI: Sử dụng transaction để xóa product và các liên kết
    tx := db.Begin()
    if err := tx.First(&product, id).Error; err != nil {
        tx.Rollback()
        if errors.Is(err, gorm.ErrRecordNotFound) {
             c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
             return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }

    // Xóa các liên kết trong bảng product_categories
    if err := tx.Model(&product).Association("Categories").Clear(); err != nil {
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear product associations"})
        return
    }

    // Xóa sản phẩm
    if err := tx.Delete(&product).Error; err != nil {
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
        return
    }

    if err := tx.Commit().Error; err != nil {
         c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Product deleted"})
}

func SearchProducts(c *gin.Context) {
    db := database.GetDB()
    var products []models.Product

    query := c.Query("q")

    if query == "" {
        c.JSON(http.StatusOK, products)
        return
    }

    searchTerm := "%" + query + "%"

    // THAY ĐỔI: Preload "Categories" (số nhiều)
    if err := db.Preload("Categories").
                  Order("created_at DESC").
                  Where("LOWER(name) LIKE LOWER(?)", searchTerm).
                  Limit(10).
                  Find(&products).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    // Bỏ vòng lặp gán CategoryName

    c.JSON(http.StatusOK, products)
}

func GetProductsByCategory(c *gin.Context) {
    db := database.GetDB()

    categoryIDStr := c.Param("id")
    categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
        return
    }

    // Kiểm tra category tồn tại
    var category models.Category
    if err := db.First(&category, categoryID).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }

    var products []models.Product
    // THAY ĐỔI: Truy vấn qua bảng trung gian
    err = db.Preload("Categories").
        Joins("JOIN product_categories ON product_categories.product_id = products.id").
        Where("product_categories.category_id = ?", categoryID).
        Order("products.created_at DESC").
        Find(&products).Error

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve products by category"})
        return
    }

    c.JSON(http.StatusOK, products)
}

func GetAllProductIDs(c *gin.Context) {
    db := database.GetDB()
    var ids []uint

    if err := db.Model(&models.Product{}).Pluck("id", &ids).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve product IDs"})
        return
    }

    c.JSON(http.StatusOK, ids)
}

func GetSitemapProducts(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var products []models.Product
        if err := db.Select("ID", "UpdatedAt").Find(&products).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve products for sitemap"})
            return
        }

        var result []gin.H
        for _, p := range products {
            result = append(result, gin.H{
                "id":        p.ID,
                "updatedAt": p.UpdatedAt,
            })
        }

        c.JSON(http.StatusOK, result)
    }
}