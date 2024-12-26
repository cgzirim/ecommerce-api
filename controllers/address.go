package controllers

import (
	"net/http"

	"github.com/cgzirim/ecommerce-api/db"
	"github.com/cgzirim/ecommerce-api/dtos"
	"github.com/cgzirim/ecommerce-api/models"
	"github.com/gin-gonic/gin"
)

// ListAddresses godoc
// @Summary List all addresses for a user
// @Description Retrieve all addresses associated with the logged in user.
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {array} dtos.AddressDetail "Successfully retrieved addresses"
// @Failure 401 {object} dtos.ErrorResponse "Unauthenticated, login is required"
// @Failure 500 {object} dtos.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /users/addresses [get]
func ListAddresses(c *gin.Context) {
	authUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{Error: "Unauthenticated, login is required"})
		return
	}

	user := authUser.(models.User)

	var addresses []models.Address
	if err := db.DB.Where("user_id = ?", user.ID).Find(&addresses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: "Failed to retrieve addresses"})
		return
	}

	c.JSON(http.StatusOK, addresses)
}

// CreateAddress godoc
// @Summary Create a new address
// @Description Create a new address for the logged in user
// @Tags User
// @Accept json
// @Produce json
// @Param address body dtos.CreateAddressRequest true "Address information"
// @Success 201 {object} dtos.AddressDetail "Address created successfully"
// @Failure 400 {object} dtos.ErrorResponse "Invalid input"
// @Failure 401 {object} dtos.ErrorResponse "Unauthenticated, login is required"
// @Failure 500 {object} dtos.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /addresses [post]
func CreateAddress(c *gin.Context) {
	var request dtos.CreateAddressRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: "Invalid input data"})
		return
	}

	authUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{Error: "Unauthenticated, login is required"})
		return
	}

	user := authUser.(models.User)

	address := models.Address{
		FirstName:     request.FirstName,
		LastName:      request.LastName,
		City:          request.City,
		Country:       request.Country,
		ZipCode:       request.ZipCode,
		StreetAddress: request.StreetAddress,
		UserID:        user.ID,
	}

	if err := db.DB.Create(&address).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: "Failed to create address"})
		return
	}

	c.JSON(http.StatusCreated, address)
}
