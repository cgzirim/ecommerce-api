package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/cgzirim/ecommerce-api/db"
	"github.com/cgzirim/ecommerce-api/dtos"
	"github.com/cgzirim/ecommerce-api/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCreateOrder(t *testing.T) {
	mockDB, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	mockDB.AutoMigrate(&models.Order{}, &models.User{}, &models.Address{}, &models.Product{}, &models.OrderItem{})

	// Override the global DB variable with the mock DB and reset it after the test
	originalDB := db.DB
	db.SetMockDB(mockDB)
	defer func() { db.DB = originalDB }()

	user := models.User{Email: "test@example.com", FirstName: "John", LastName: "Doe", Role: "customer", Password: "password"}
	mockDB.Create(&user)

	address := models.Address{FirstName: "John", LastName: "Doe", City: "CityA", Country: "CountryA", ZipCode: "12345", StreetAddress: "Street 1", UserID: user.ID}
	mockDB.Create(&address)

	product := models.Product{Name: "Product A", Price: 10.0}
	mockDB.Create(&product)

	t.Run("Successfully creates order", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		router := gin.Default()
		router.POST("/orders", func(c *gin.Context) {
			c.Set("user", user)
			CreateOrder(c)
		})

		orderRequest := dtos.CreateOrderRequest{
			AddressID: address.ID,
			OrderItems: []dtos.OrderItemRequest{
				{ProductID: product.ID, Quantity: 2},
			},
		}
		body, _ := json.Marshal(orderRequest)
		req, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)

		var createdOrder models.Order
		err := json.Unmarshal(rec.Body.Bytes(), &createdOrder)
		assert.NoError(t, err)

		var reloadedOrder models.Order
		mockDB.Preload("User").Preload("Address").Preload("OrderItems").First(&reloadedOrder, createdOrder.ID)

		assert.Equal(t, user.ID, reloadedOrder.UserID)
		assert.Equal(t, address.ID, reloadedOrder.AddressID)
		assert.Equal(t, models.OrderStatusPending, reloadedOrder.Status)
		assert.Equal(t, 20.0, reloadedOrder.Total)
		assert.Len(t, reloadedOrder.OrderItems, 1)
		assert.Equal(t, product.ID, reloadedOrder.OrderItems[0].ProductID)
		assert.Equal(t, 2, reloadedOrder.OrderItems[0].Quantity)
	})

	t.Run("Fails with invalid input data", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		router := gin.Default()
		router.POST("/orders", func(c *gin.Context) {
			c.Set("user", user)
			CreateOrder(c)
		})

		invalidRequest := map[string]interface{}{
			"address_id": address.ID,
			"order_items": []map[string]interface{}{
				{"product_id": product.ID, "quantity": 0}, // Invalid quantity
			},
		}
		body, _ := json.Marshal(invalidRequest)
		req, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Quantity must be greater than 0 for product ID: 1", response["error"])
	})

	t.Run("Fails when user is unauthenticated", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		router := gin.Default()
		router.POST("/orders", CreateOrder)

		orderRequest := dtos.CreateOrderRequest{
			AddressID: address.ID,
			OrderItems: []dtos.OrderItemRequest{
				{ProductID: product.ID, Quantity: 2},
			},
		}
		body, _ := json.Marshal(orderRequest)
		req, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Unauthenticated, login is required", response.Error)
	})
}

func TestCancelOrder(t *testing.T) {
	mockDB, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	mockDB.AutoMigrate(&models.Order{}, &models.User{}, &models.Address{}, &models.Product{}, &models.OrderItem{})

	originalDB := db.DB
	db.SetMockDB(mockDB)
	defer func() { db.DB = originalDB }()

	user := models.User{Email: "test@example.com", FirstName: "John", LastName: "Doe", Role: "customer", Password: "password"}
	mockDB.Create(&user)

	address := models.Address{FirstName: "John", LastName: "Doe", City: "CityA", Country: "CountryA", ZipCode: "12345", StreetAddress: "Street 1", UserID: user.ID}
	mockDB.Create(&address)

	order := models.Order{
		UserID:    user.ID,
		AddressID: address.ID,
		Total:     20.0,
		Status:    models.OrderStatusPending,
	}
	mockDB.Create(&order)

	t.Run("Successfully cancels order", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		router := gin.Default()
		router.PATCH("/orders/:id/cancel", func(c *gin.Context) {
			c.Set("user", user)
			CancelOrder(c)
		})

		req, _ := http.NewRequest("PATCH", "/orders/"+strconv.Itoa(int(order.ID))+"/cancel", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var cancelledOrder models.Order
		err := json.Unmarshal(rec.Body.Bytes(), &cancelledOrder)
		assert.NoError(t, err)

		assert.Equal(t, models.OrderStatusCancelled, cancelledOrder.Status)
	})

	t.Run("Fails when user is unauthenticated", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		router := gin.Default()
		router.PATCH("/orders/:id/cancel", CancelOrder)

		req, _ := http.NewRequest("PATCH", "/orders/"+strconv.Itoa(int(order.ID))+"/cancel", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Unauthenticated, login is required", response.Error)
	})

	t.Run("Fails when order is not in pending status", func(t *testing.T) {
		order.Status = models.OrderStatusCompleted
		mockDB.Save(&order)

		gin.SetMode(gin.TestMode)

		router := gin.Default()
		router.PATCH("/orders/:id/cancel", func(c *gin.Context) {
			c.Set("user", user)
			CancelOrder(c)
		})

		req, _ := http.NewRequest("PATCH", "/orders/"+strconv.Itoa(int(order.ID))+"/cancel", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Order cannot be cancelled, it is not in pending state", response.Error)
	})
}

func TestUpdateOrderStatus(t *testing.T) {
	mockDB, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	mockDB.AutoMigrate(&models.Order{}, &models.User{}, &models.Address{}, &models.Product{}, &models.OrderItem{})

	originalDB := db.DB
	db.SetMockDB(mockDB)
	defer func() { db.DB = originalDB }()

	admin := models.User{Email: "admin@example.com", FirstName: "Admin", LastName: "User", Role: "admin", Password: "password"}
	mockDB.Create(&admin)

	user := models.User{Email: "test@example.com", FirstName: "John", LastName: "Doe", Role: "customer", Password: "password"}
	mockDB.Create(&user)

	address := models.Address{FirstName: "John", LastName: "Doe", City: "CityA", Country: "CountryA", ZipCode: "12345", StreetAddress: "Street 1", UserID: user.ID}
	mockDB.Create(&address)

	order := models.Order{
		UserID:    user.ID,
		AddressID: address.ID,
		Total:     20.0,
		Status:    models.OrderStatusPending,
	}
	mockDB.Create(&order)

	t.Run("Successfully updates order status", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		router := gin.Default()
		router.PATCH("/orders/:id/status", func(c *gin.Context) {
			c.Set("user", admin)
			UpdateOrderStatus(c)
		})

		updateRequest := dtos.UpdateOrderStatusRequest{
			Status: models.OrderStatusCompleted,
		}
		body, _ := json.Marshal(updateRequest)
		req, _ := http.NewRequest("PATCH", "/orders/"+strconv.Itoa(int(order.ID))+"/status", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var updatedOrder models.Order
		err := json.Unmarshal(rec.Body.Bytes(), &updatedOrder)
		assert.NoError(t, err)

		assert.Equal(t, models.OrderStatusCompleted, updatedOrder.Status)
	})

	t.Run("Fails with invalid order status", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		router := gin.Default()
		router.PATCH("/orders/:id/status", func(c *gin.Context) {
			c.Set("user", admin)
			UpdateOrderStatus(c)
		})

		updateRequest := dtos.UpdateOrderStatusRequest{
			Status: "invalid_status",
		}
		body, _ := json.Marshal(updateRequest)
		req, _ := http.NewRequest("PATCH", "/orders/"+strconv.Itoa(int(order.ID))+"/status", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid order status", response.Error)
	})

	t.Run("Fails when user is unauthenticated", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		router := gin.Default()
		router.PATCH("/orders/:id/status", UpdateOrderStatus)

		updateRequest := dtos.UpdateOrderStatusRequest{
			Status: models.OrderStatusCompleted,
		}
		body, _ := json.Marshal(updateRequest)
		req, _ := http.NewRequest("PATCH", "/orders/"+strconv.Itoa(int(order.ID))+"/status", bytes.NewBuffer(body))
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
		gin.SetMode(gin.TestMode)

		router := gin.Default()
		router.PATCH("/orders/:id/status", func(c *gin.Context) {
			c.Set("user", user)
			UpdateOrderStatus(c)
		})

		updateRequest := dtos.UpdateOrderStatusRequest{
			Status: models.OrderStatusCompleted,
		}
		body, _ := json.Marshal(updateRequest)
		req, _ := http.NewRequest("PATCH", "/orders/"+strconv.Itoa(int(order.ID))+"/status", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Unauthorized, only admins can update order status", response.Error)
	})

}
