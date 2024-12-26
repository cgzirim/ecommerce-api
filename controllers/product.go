package controllers

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/cgzirim/ecommerce-api/db"
	"github.com/cgzirim/ecommerce-api/dtos"
	"github.com/cgzirim/ecommerce-api/models"

	"github.com/gin-gonic/gin"
)

// ListProducts godoc
// @Summary Retrieve a paginated list of products
// @Description Retrieve a paginated list of products with the ability to specify page and page size.
// @Tags Product
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Number of products per page" default(10)
// @Success 200 {object} dtos.ProductListResponse "Successfully retrieved the paginated list of products"
// @Failure 400 {object} dtos.ErrorResponse "Invalid page number or pageSize"
// @Failure 500 {object} dtos.ErrorResponse "Internal server error"
// @Router /products [get]
func ListProducts(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: "Invalid page number"})
		return
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if err != nil || pageSize <= 0 {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: "Invalid pageSize number"})
		return
	}

	offset := (page - 1) * pageSize

	var products []models.Product

	result := db.DB.Limit(pageSize).Offset(offset).Find(&products)
	if result.Error != nil {
		log.Printf("Failed to retrieve products: %v", result.Error)
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: result.Error.Error()})
		return
	}

	var totalProducts int64
	db.DB.Model(&models.Product{}).Count(&totalProducts)

	c.JSON(http.StatusOK, gin.H{
		"page":        page,
		"page_size":   pageSize,
		"total_count": totalProducts,
		"total_pages": int(math.Ceil(float64(totalProducts) / float64(pageSize))),
		"products":    products,
	})
}

// GetProductByID godoc
// @Summary Retrieve a product by ID
// @Description Retrieve a product by its unique ID.
// @Tags Product
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} dtos.ProductDetail
// @Failure 400 {object} dtos.ErrorResponse "Invalid product ID"
// @Failure 404 {object} dtos.ErrorResponse "Product not found"
// @Failure 500 {object} dtos.ErrorResponse "Internal server error"
// @Router /products/{id} [get]
func GetProductByID(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: "Invalid product ID"})
		return
	}

	var product models.Product
	result := db.DB.First(&product, productID)
	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			c.JSON(http.StatusNotFound, dtos.ErrorResponse{Error: "Product not found"})
		} else {
			log.Printf("Failed to retrieve product with ID (%v): %v", productID, result.Error)
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: result.Error.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, product)
}

// CreateProduct godoc
// @Summary Create a new product
// @Description Allows an admin to create a new product
// @Tags Product
// @Accept json
// @Produce json
// @Param input body dtos.CreateProductRequest true "Product information"
// @Success 201 {object} models.Product "Product created successfully"
// @Failure 400 {object} dtos.ErrorResponse "Invalid input data"
// @Failure 401 {object} dtos.ErrorResponse "Unauthenticated, login is required"
// @Failure 403 {object} dtos.ErrorResponse "Unauthorized access, only admins can create products"
// @Failure 500 {object} dtos.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /products [post]
func CreateProduct(c *gin.Context) {
	var req dtos.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationErrors(err, c)
		return
	}

	authUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{Error: "Unauthenticated, login is required"})
		return
	}

	user := authUser.(models.User)

	if !user.IsAdmin() {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized access, only admins can create products"})
		return
	}

	product := models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		Category:    req.Category,
	}

	if err := db.DB.Create(&product).Error; err != nil {
		log.Printf("Failed to create product: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create product: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, product)
}

// UpdateProduct godoc
// @Summary Fully update an existing product
// @Description Allows an admin to fully update all fields of an existing product by providing the product ID and new data.
// @Tags Product
// @Accept json
// @Produce json
//
// @Param id path int true "Product ID"
// @Param input body dtos.CreateProductRequest true "Product data to update"
// @Success 200 {object} models.Product "Updated product details"
// @Failure 400 {object} dtos.ErrorResponse "Invalid product ID or request payload"
// @Failure 401 {object} dtos.ErrorResponse "Unauthenticated, login is required"
// @Failure 403 {object} dtos.ErrorResponse "Unauthorized access, only admins can update products"
// @Failure 404 {object} dtos.ErrorResponse "Product not found"
// @Failure 500 {object} dtos.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /products/{id} [put]
func UpdateProduct(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: "Invalid product ID"})
		return
	}

	authUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{Error: "Unauthenticated, login is required"})
		return
	}

	user := authUser.(models.User)

	if !user.IsAdmin() {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized access, only admins can update products"})
		return
	}

	var product models.Product
	result := db.DB.First(&product, productID)
	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			c.JSON(http.StatusNotFound, dtos.ErrorResponse{Error: "Product not found"})
		} else {
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: result.Error.Error()})
		}
		return
	}

	var req dtos.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationErrors(err, c)
		return
	}

	if err := db.DB.Model(product).Updates(req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: "Failed to update product"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// PatchProduct godoc
// @Summary Partially update an existing product
// @Description Allows an admin to update specific fields of an existing product by providing the product ID and the updated data.
// @Tags Product
// @Accept json
// @Produce json
//
// @Param id path int true "Product ID"
// @Param input body dtos.PatchProductRequest true "Product data to patch"
// @Success 200 {object} models.Product "Updated product details"
// @Failure 400 {object} dtos.ErrorResponse "Invalid product ID or request payload"
// @Failure 401 {object} dtos.ErrorResponse "Unauthenticated, login is required"
// @Failure 403 {object} dtos.ErrorResponse "Unauthorized access, only admins can patch products"
// @Failure 404 {object} dtos.ErrorResponse "Product not found"
// @Failure 500 {object} dtos.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /products/{id} [patch]
func PatchProduct(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: "Invalid product ID"})
		return
	}

	authUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{Error: "Unauthenticated, login is required"})
		return
	}

	user := authUser.(models.User)

	if !user.IsAdmin() {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized access, only admins can patch products"})
		return
	}

	var product models.Product
	result := db.DB.First(&product, productID)
	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			c.JSON(http.StatusNotFound, dtos.ErrorResponse{Error: "Product not found"})
		} else {
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: result.Error.Error()})
		}
		return
	}

	var req dtos.PatchProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationErrors(err, c)
		return
	}

	if err := db.DB.Model(&product).Updates(req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}

// DeleteProduct deletes a product by its ID
// @Summary Delete a product
// @Description Allows an admin to delete a product by its ID
// @Tags Product
// @Param id path int true "Product ID"
// @Success 204 "Product deleted successfully"
// @Failure 400 {object} dtos.ErrorResponse "Invalid product ID"
// @Failure 401 {object} dtos.ErrorResponse "Unauthenticated, login is required"
// @Failure 403 {object} dtos.ErrorResponse "Unauthorized access, only admins can delete products"
// @Failure 404 {object} dtos.ErrorResponse "Product not found"
// @Failure 500 {object} dtos.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /products/{id} [delete]
func DeleteProduct(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: "Invalid product ID"})
		return
	}

	authUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{Error: "Unauthenticated, login is required"})
		return
	}

	user := authUser.(models.User)

	if !user.IsAdmin() {
		c.JSON(http.StatusForbidden, dtos.ErrorResponse{Error: "Unauthorized access, only admins can delete products"})
		return
	}

	var product models.Product
	result := db.DB.First(&product, productID)
	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			c.JSON(http.StatusNotFound, dtos.ErrorResponse{Error: "Product not found"})
		} else {
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: result.Error.Error()})
		}
		return
	}

	if err := db.DB.Delete(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: fmt.Sprintf("Failed to delete product: %v", err)})
		return
	}

	c.Status(http.StatusNoContent)
}
