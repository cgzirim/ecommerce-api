package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cgzirim/ecommerce-api/db"
	"github.com/cgzirim/ecommerce-api/dtos"
	"github.com/cgzirim/ecommerce-api/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCreateProduct(t *testing.T) {
	mockDB, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	mockDB.AutoMigrate(&models.Product{}, &models.User{})

	originalDB := db.DB
	db.SetMockDB(mockDB)
	defer func() { db.DB = originalDB }()

	admin := models.User{Email: "admin@example.com", FirstName: "Admin", LastName: "User", Role: "admin", Password: "password"}
	mockDB.Create(&admin)

	t.Run("Successfully creates product", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		router := gin.Default()
		router.POST("/products", func(c *gin.Context) {
			c.Set("user", admin)
			CreateProduct(c)
		})

		productRequest := dtos.CreateProductRequest{
			Name:        "Product A",
			Description: "Description of Product A",
			Price:       10.0,
			Stock:       100,
			Category:    "Category A",
		}
		body, _ := json.Marshal(productRequest)
		req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)

		var createdProduct models.Product
		err := json.Unmarshal(rec.Body.Bytes(), &createdProduct)
		assert.NoError(t, err)

		assert.Equal(t, productRequest.Name, createdProduct.Name)
		assert.Equal(t, productRequest.Description, createdProduct.Description)
		assert.Equal(t, productRequest.Price, createdProduct.Price)
		assert.Equal(t, productRequest.Stock, createdProduct.Stock)
		assert.Equal(t, productRequest.Category, createdProduct.Category)
	})

	t.Run("Fails with invalid input data", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		router := gin.Default()
		router.POST("/products", func(c *gin.Context) {
			c.Set("user", admin)
			CreateProduct(c)
		})

		invalidRequest := map[string]interface{}{
			"name": "Product A",
			// Missing required fields
		}
		body, _ := json.Marshal(invalidRequest)
		req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Fails when user is unauthenticated", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		router := gin.Default()
		router.POST("/products", CreateProduct)

		productRequest := dtos.CreateProductRequest{
			Name:        "Product A",
			Description: "Description of Product A",
			Price:       10.0,
			Stock:       100,
			Category:    "Category A",
		}
		body, _ := json.Marshal(productRequest)
		req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Unauthenticated, login is required", response.Error)
	})

	t.Run("Fails when user is not an admin", func(t *testing.T) {
		user := models.User{Email: "user@example.com", FirstName: "User", LastName: "Doe", Role: "customer", Password: "password"}
		mockDB.Create(&user)

		gin.SetMode(gin.TestMode)

		router := gin.Default()
		router.POST("/products", func(c *gin.Context) {
			c.Set("user", user)
			CreateProduct(c)
		})

		productRequest := dtos.CreateProductRequest{
			Name:        "Product A",
			Description: "Description of Product A",
			Price:       10.0,
			Stock:       100,
			Category:    "Category A",
		}
		body, _ := json.Marshal(productRequest)
		req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Unauthorized access, only admins can create products", response.Error)
	})
}
