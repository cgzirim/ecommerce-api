package controllers

import (
	"net/http"
	"strings"
	"time"

	"log"

	"github.com/cgzirim/ecommerce-api/db"
	"github.com/cgzirim/ecommerce-api/dtos"
	"github.com/cgzirim/ecommerce-api/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

const ADMIN_SECRET_KEY = "admin123"
const LOGIN_ON_REGISTRATION = true

// RegisterCustomer godoc
// @Summary Register a new customer
// @Description Allows a user to register as a customer by providing necessary details.
// @Tags Auth
// @Accept json
// @Produce json
//
// @Param input body dtos.CustomerRegistrationRequest true "Customer registration details"
//
// @Success 200 {object} dtos.RegistrationSuccessResponse "Successfully registered customer"
// @Failure 400 {object} dtos.ErrorResponse "Validation error or mismatched passwords"
// @Failure 500 {object} dtos.ErrorResponse "Internal server error"
// @Router /register [post]
func RegisterCustomer(c *gin.Context) {
	var req dtos.CustomerRegistrationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationErrors(err, c)
		return
	}

	if req.Password != req.PasswordConfirm {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Passwords do not match"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Password:  string(hashedPassword),
		Role:      models.RoleCustomer,
	}

	if err := db.DB.Create(&user).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint \"uni_users_email\"") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User with this email already exists."})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var accessToken, refreshToken string

	if LOGIN_ON_REGISTRATION {
		accessToken, refreshToken, err = models.GenerateJwtTokens(&user)
		if err != nil {
			log.Printf("failed to generate JWT token for user %v: %v", user.Email, err)
		}
	}

	response := dtos.RegistrationSuccessResponse{
		Msg:          "Account registered successfully",
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	c.JSON(http.StatusOK, response)
}

// RegisterAdmin godoc
// @Summary Register a new admin
// @Description Allows a user to register as an admin by providing the necessary details, including a secret key for authentication (```secret_key: admin123```).
// @Tags Auth
// @Accept json
// @Produce json
//
// @Param input body dtos.AdminRegistrationRequest true "Admin registration details"
//
// @Success 200 {object} dtos.RegistrationSuccessResponse "Account registered successfully"
// @Failure 400 {object} dtos.ErrorResponse "Validation error or mismatched passwords"
// @Failure 500 {object} dtos.ErrorResponse "Internal server error"
// @Router /register/admin [post]
func RegisterAdmin(c *gin.Context) {
	var req dtos.AdminRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationErrors(err, c)
		return
	}

	if req.Password != req.PasswordConfirm {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Passwords do not match"})
		return
	}

	if req.SecretKey != ADMIN_SECRET_KEY {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid secret key for admin registration"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Password:  string(hashedPassword),
		Role:      models.RoleAdmin,
	}

	if err := db.DB.Create(&user).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint \"uni_users_email\"") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User with this email already exists."})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var accessToken, refreshToken string

	if LOGIN_ON_REGISTRATION {
		accessToken, refreshToken, err = models.GenerateJwtTokens(&user)
		if err != nil {
			log.Printf("failed to generate JWT token for user %v: %v", user.Email, err)
		}
	}

	response := dtos.RegistrationSuccessResponse{
		Msg:          "Registration successful",
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	c.JSON(http.StatusOK, response)
}

// LoginUser godoc
// @Summary User login
// @Description Allows a user to login by providing email and password.
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body dtos.LoginRequest true "Login details"
// @Success 200 {object} dtos.LoginSuccessResponse "Login successful"
// @Failure 400 {object} dtos.ErrorResponse "Validation error"
// @Failure 401 {object} dtos.ErrorResponse "Invalid credentials"
// @Failure 500 {object} dtos.ErrorResponse "Internal server error"
// @Router /login [post]
func LoginUser(c *gin.Context) {
	var req dtos.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationErrors(err, c)
		return
	}

	var user models.User
	if err := db.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := user.IsValidPassword(req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	accessToken, refreshToken, err := models.GenerateJwtTokens(&user)
	if err != nil {
		log.Printf("failed to generate JWT token for user %v: %v", user.Email, err)
	}

	response := dtos.LoginSuccessResponse{
		Msg:          "Login successful",
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	// Update LastLogin to current time
	user.LastLogin = time.Now()
	if err := db.DB.Save(&user).Error; err != nil {
		log.Printf("Failed to update LastLogin for user %v: %v", user.Email, err)
	}

	c.JSON(http.StatusOK, response)
}
