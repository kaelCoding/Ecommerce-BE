package handlers

import (
	"errors" // THÊM IMPORT
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kaelCoding/toyBE/internal/database"
	"github.com/kaelCoding/toyBE/internal/models"
	"gorm.io/gorm"
)

func AddCategory(c *gin.Context) {
    var category models.Category
    if err := c.ShouldBindJSON(&category); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if err := database.DB.Create(&category).Error; err != nil {
        c.JSON(http.StatusConflict, gin.H{"error": "Category name already exists"})
        return
    }

    c.JSON(http.StatusCreated, category)
}

func GetCategory(c *gin.Context) {
    var category []models.Category
    // THAY ĐỔI: Tải kèm (Preload) danh sách sản phẩm
    if err := database.DB.Preload("Products").Order("id ASC").Find(&category).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve category"})
        return
    }

    c.JSON(http.StatusOK, category)
}

func GetCategoryByID(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
        return
    }

    var category models.Category
    // THAY ĐỔI: Tải kèm (Preload) danh sách sản phẩm
    if err := database.DB.Preload("Products").First(&category, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) { // Sửa: gorm.ErrRecordNotFound
            c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }

    c.JSON(http.StatusOK, category)
}

func UpdateCategory(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
        return
    }

    var category models.Category
    if err := database.DB.First(&category, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) { // Sửa: gorm.ErrRecordNotFound
             c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
             return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }

    var input models.Category
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    category.Name = input.Name
    category.Description = input.Description

    if err := database.DB.Save(&category).Error; err != nil {
        c.JSON(http.StatusConflict, gin.H{"error": "Category name already exists"})
        return
    }
    
    c.JSON(http.StatusOK, category)
}

func DeleteCategory(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
        return
    }
    
    db := database.GetDB()
    tx := db.Begin()

    var category models.Category
    if err := tx.First(&category, id).Error; err != nil {
        tx.Rollback()
        if errors.Is(err, gorm.ErrRecordNotFound) {
             c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
             return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }

    // THAY ĐỔI: Xóa các liên kết trong bảng product_categories trước
    if err := tx.Model(&category).Association("Products").Clear(); err != nil {
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear category associations"})
        return
    }

    // Xóa vĩnh viễn (Unscoped) danh mục
    if err := tx.Unscoped().Delete(&category).Error; err != nil {
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})
        return
    }

    if err := tx.Commit().Error; err != nil {
         c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Category deleted permanently"})
}

func toSlug(s string) string {
    s = strings.ToLower(s)
    s = strings.ReplaceAll(s, " ", "-")
    return s
}

func GetSitemapCategories(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var categories []models.Category
        
        if err := db.Select("ID", "Name", "UpdatedAt").Find(&categories).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve categories for sitemap"})
            return
        }

        var result []gin.H
        for _, cat := range categories {
            result = append(result, gin.H{
                "slug":      toSlug(cat.Name),
                "updatedAt": cat.UpdatedAt,
            })
        }

        c.JSON(http.StatusOK, result)
    }
}