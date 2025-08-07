package handlers

import (
    "net/http"
    "strconv"
    
    "github.com/gin-gonic/gin"
	"github.com/kaelCoding/toyBE/internal/models"
	"github.com/kaelCoding/toyBE/internal/database"
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
    if err := database.DB.Order("id ASC").Find(&category).Error; err != nil {
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
    if err := database.DB.First(&category, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
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
        c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
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
    
    if err := database.DB.Unscoped().Delete(&models.Category{}, id).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Category deleted permanently"})
}