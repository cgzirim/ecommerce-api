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
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRegisterCustomer(t *testing.T) {
	mockDB, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	mockDB.AutoMigrate(&models.User{})

	originalDB := db.DB
	db.SetMockDB(mockDB)
	defer func() { db.DB = originalDB }()

	t.Run("Successfully registers customer", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		router := gin.Default()
		router.POST("/register/customer", RegisterCustomer)

		customerRequest := dtos.CustomerRegistrationRequest{
			Email:           "customer@example.com",
			FirstName:       "John",
			LastName:        "Doe",
			Password:        "password",
			PasswordConfirm: "password",
		}
		body, _ := json.Marshal(customerRequest)
		req, _ := http.NewRequest("POST", "/register/customer", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)

		var response dtos.RegistrationSuccessResponse
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, customerRequest.Email, response.User.Email)
		assert.Equal(t, customerRequest.FirstName, response.User.FirstName)
		assert.Equal(t, customerRequest.LastName, response.User.LastName)
		assert.Equal(t, models.RoleCustomer, response.User.Role)
	})

	t.Run("Fails with invalid input data", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		router := gin.Default()
		router.POST("/register/customer", RegisterCustomer)

		invalidRequest := map[string]interface{}{
			"email": "customer@example.com",
			// Missing required fields
		}
		body, _ := json.Marshal(invalidRequest)
		req, _ := http.NewRequest("POST", "/register/customer", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Fails when passwords do not match", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		router := gin.Default()
		router.POST("/register/customer", RegisterCustomer)

		customerRequest := dtos.CustomerRegistrationRequest{
			Email:           "customer@example.com",
			FirstName:       "John",
			LastName:        "Doe",
			Password:        "password",
			PasswordConfirm: "differentpassword",
		}
		body, _ := json.Marshal(customerRequest)
		req, _ := http.NewRequest("POST", "/register/customer", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Passwords do not match", response.Error)
	})
}

func TestRegisterAdmin(t *testing.T) {
	mockDB, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	mockDB.AutoMigrate(&models.User{})

	originalDB := db.DB
	db.SetMockDB(mockDB)
	defer func() { db.DB = originalDB }()

	t.Run("Successfully registers admin", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		router := gin.Default()
		router.POST("/register/admin", RegisterAdmin)

		adminRequest := dtos.AdminRegistrationRequest{
			Email:           "admin@example.com",
			FirstName:       "Admin",
			LastName:        "User",
			Password:        "password",
			PasswordConfirm: "password",
			SecretKey:       ADMIN_SECRET_KEY,
		}
		body, _ := json.Marshal(adminRequest)
		req, _ := http.NewRequest("POST", "/register/admin", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)

		var response dtos.RegistrationSuccessResponse
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, adminRequest.Email, response.User.Email)
		assert.Equal(t, adminRequest.FirstName, response.User.FirstName)
		assert.Equal(t, adminRequest.LastName, response.User.LastName)
		assert.Equal(t, models.RoleAdmin, response.User.Role)
	})

	t.Run("Fails with incorrect secret key", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		router := gin.Default()
		router.POST("/register/admin", RegisterAdmin)

		adminRequest := dtos.AdminRegistrationRequest{
			Email:           "admin@example.com",
			FirstName:       "Admin",
			LastName:        "User",
			Password:        "password",
			PasswordConfirm: "password",
			SecretKey:       "wrongsecret#$%",
		}
		body, _ := json.Marshal(adminRequest)
		req, _ := http.NewRequest("POST", "/register/admin", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid secret key for admin registration", response.Error)
	})
}

func TestLogin(t *testing.T) {
	mockDB, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	mockDB.AutoMigrate(&models.User{})

	originalDB := db.DB
	db.SetMockDB(mockDB)
	defer func() { db.DB = originalDB }()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := models.User{Email: "user@example.com", FirstName: "John", LastName: "Doe", Role: "customer", Password: string(hashedPassword)}
	mockDB.Create(&user)

	t.Run("Successfully logs in user", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		router := gin.Default()
		router.POST("/login", LoginUser)

		loginRequest := dtos.LoginRequest{
			Email:    "user@example.com",
			Password: "password",
		}
		body, _ := json.Marshal(loginRequest)
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response dtos.LoginSuccessResponse
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, user.Email, response.User.Email)
		assert.Equal(t, user.FirstName, response.User.FirstName)
		assert.Equal(t, user.LastName, response.User.LastName)
		assert.NotEmpty(t, response.AccessToken)
		assert.NotEmpty(t, response.RefreshToken)
	})

	t.Run("Fails with invalid credentials", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		router := gin.Default()
		router.POST("/login", LoginUser)

		loginRequest := dtos.LoginRequest{
			Email:    "user@example.com",
			Password: "wrongpassword",
		}
		body, _ := json.Marshal(loginRequest)
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid credentials", response.Error)
	})
}
