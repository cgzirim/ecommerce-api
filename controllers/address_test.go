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

func TestListAddresses(t *testing.T) {
	mockDB, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	mockDB.AutoMigrate(&models.Address{}, &models.User{})

	// Override the global DB variable with the mock DB and reset it after the test
	originalDB := db.DB
	defer func() { db.DB = originalDB }()
	db.SetMockDB(mockDB)

	user := models.User{Email: "test@example.com", FirstName: "John", LastName: "Doe", Role: "customer", Password: "password"}
	mockDB.Create(&user)

	addresses := []models.Address{
		{FirstName: "John", LastName: "Doe", City: "CityA", Country: "CountryA", ZipCode: "12345", StreetAddress: "Street 1", UserID: user.ID},
		{FirstName: "Jane", LastName: "Doe", City: "CityB", Country: "CountryB", ZipCode: "67890", StreetAddress: "Street 2", UserID: user.ID},
	}
	for _, addr := range addresses {
		mockDB.Create(&addr)
	}

	t.Run("Successfully retrieves addresses", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		router := gin.Default()
		router.GET("/users/addresses", func(c *gin.Context) {
			c.Set("user", user)
			ListAddresses(c)
		})

		req, _ := http.NewRequest("GET", "/users/addresses", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var retrievedAddresses []models.Address
		err := json.Unmarshal(rec.Body.Bytes(), &retrievedAddresses)
		assert.NoError(t, err)

		assert.Equal(t, len(addresses), len(retrievedAddresses))
		for i, addr := range addresses {
			assert.Equal(t, addr.FirstName, retrievedAddresses[i].FirstName)
			assert.Equal(t, addr.LastName, retrievedAddresses[i].LastName)
			assert.Equal(t, addr.City, retrievedAddresses[i].City)
			assert.Equal(t, addr.Country, retrievedAddresses[i].Country)
			assert.Equal(t, addr.ZipCode, retrievedAddresses[i].ZipCode)
			assert.Equal(t, addr.StreetAddress, retrievedAddresses[i].StreetAddress)
		}
	})

	t.Run("Fails when user is unauthenticated", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		// define route without setting the user context
		router := gin.Default()
		router.GET("/users/addresses", func(c *gin.Context) {
			ListAddresses(c)
		})

		req, _ := http.NewRequest(http.MethodGet, "/users/addresses", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code, "Expected status code 401 for unauthenticated access")

		var response dtos.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "Response body should be parsable as an ErrorResponse")
		assert.Equal(t, "Unauthenticated, login is required", response.Error, "Error message should match expected text")
	})

}

func TestCreateAddress(t *testing.T) {
	mockDB, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	mockDB.AutoMigrate(&models.Address{}, &models.User{})

	// Override the global DB variable with the mock DB and reset it after the test
	originalDB := db.DB
	defer func() { db.DB = originalDB }()
	db.SetMockDB(mockDB)

	user := models.User{Email: "test@example.com", FirstName: "John", LastName: "Doe", Role: "customer", Password: "password"}
	mockDB.Create(&user)

	t.Run("Successfully creates address", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		router := gin.Default()
		router.POST("/users/addresses", func(c *gin.Context) {
			c.Set("user", user)
			CreateAddress(c)
		})

		addressRequest := dtos.CreateAddressRequest{
			FirstName:     "John",
			LastName:      "Doe",
			City:          "CityA",
			Country:       "CountryA",
			ZipCode:       "12345",
			StreetAddress: "Street 1",
		}
		body, _ := json.Marshal(addressRequest)
		req, _ := http.NewRequest("POST", "/users/addresses", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)

		var createdAddress models.Address
		err := json.Unmarshal(rec.Body.Bytes(), &createdAddress)
		assert.NoError(t, err)

		assert.Equal(t, addressRequest.FirstName, createdAddress.FirstName)
		assert.Equal(t, addressRequest.LastName, createdAddress.LastName)
		assert.Equal(t, addressRequest.City, createdAddress.City)
		assert.Equal(t, addressRequest.Country, createdAddress.Country)
		assert.Equal(t, addressRequest.ZipCode, createdAddress.ZipCode)
		assert.Equal(t, addressRequest.StreetAddress, createdAddress.StreetAddress)
		assert.Equal(t, user.ID, createdAddress.UserID)
	})

	t.Run("Fails with invalid input data", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		router := gin.Default()
		router.POST("/users/addresses", func(c *gin.Context) {
			c.Set("user", user)
			CreateAddress(c)
		})

		invalidRequest := map[string]string{
			"first_name": "John",
			// Missing required fields
		}
		body, _ := json.Marshal(invalidRequest)
		req, _ := http.NewRequest("POST", "/users/addresses", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var response dtos.ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid input data", response.Error)
	})

	t.Run("Fails when user is unauthenticated", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		router := gin.Default()
		router.POST("/users/addresses", CreateAddress)

		addressRequest := dtos.CreateAddressRequest{
			FirstName:     "John",
			LastName:      "Doe",
			City:          "CityA",
			Country:       "CountryA",
			ZipCode:       "12345",
			StreetAddress: "Street 1",
		}
		body, _ := json.Marshal(addressRequest)
		req, _ := http.NewRequest("POST", "/users/addresses", bytes.NewBuffer(body))
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
