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
    "github.com/kaelCoding/toyBE/internal/pkg/r2"
)

func createProductResponse(db *gorm.DB, product *models.Product) (map[string]interface{}, error) {
    if err := db.Preload("Category").First(&product, product.ID).Error; err != nil {
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
        "category_id":   product.CategoryID,
        "category_name": product.Category.Name, 
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
    
    var catID uint
    fmt.Sscanf(categoryID, "%d", &catID)
    
    product := models.Product{
        Name:        name,
        Description: description,
        Price:       price,
        CategoryID:  catID,
        ImageURLs:   imageURLsJSON,
    }

    var category models.Category
    if err := db.First(&category, product.CategoryID).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    if err := db.Create(&product).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product: " + err.Error()})
        return
    }
    
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
    if err := db.Where("category_id = ?", id).Limit(limit).Order("created_at DESC").Preload("Category").Find(&products).Error; err != nil {
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

    catID, _ := strconv.Atoi(c.PostForm("category_id"))
    existingProduct.CategoryID = uint(catID)

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

    if err := db.Save(&existingProduct).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
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

func SearchProducts(c *gin.Context) {
    db := database.GetDB()
    var products []models.Product

    query := c.Query("q")

    if query == "" {
        c.JSON(http.StatusOK, products)
        return
    }

    searchTerm := "%" + query + "%"

    if err := db.Preload("Category").
                  Order("created_at DESC").
                  Where("LOWER(name) LIKE LOWER(?)", searchTerm).
                  Limit(10).
                  Find(&products).Error; err != nil {
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

func GetProductsByCategory(c *gin.Context) {
    db := database.GetDB()

    categoryIDStr := c.Param("id")
    categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
        return
    }

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
    if err := db.Preload("Category").Where("category_id = ?", categoryID).Order("created_at DESC").Find(&products).Error; err != nil {
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