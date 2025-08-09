package handlers

import (
    "net/http"
    "strconv"
    "errors"
    "gorm.io/gorm"
    "fmt"
    "time"
    "encoding/json"

    
    "github.com/gin-gonic/gin"
	"github.com/kaelCoding/toyBE/internal/models"
	"github.com/kaelCoding/toyBE/internal/database"
    "github.com/kaelCoding/toyBE/internal/pkg/cloudinary"
)

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

		publicID := fmt.Sprintf("product_%d", time.Now().UnixNano())

		fileURL, err := cloudinary.UploadToCloudinary(fileContent, "products", publicID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to upload file to Cloudinary: " + err.Error()})
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
				fileContent, err := file.Open()
				if err != nil {
			        return
				}
				defer fileContent.Close()

				publicID := fmt.Sprintf("product_update_%d", time.Now().UnixNano())
				fileURL, err := cloudinary.UploadToCloudinary(fileContent, "products", publicID)
				if err != nil {
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