package dtos

import "github.com/cgzirim/ecommerce-api/models"

// LoginRequest represents the expected request body for logging in a user
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"securepassword"`
}

// LoginSuccessResponse represents the response body for a successful login
type LoginSuccessResponse struct {
	User         models.User `json:"user"`
	AccessToken  string      `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string      `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	Msg          string      `json:"msg" example:"Login successful"`
}

// AdminRegistrationRequest represents the expected request body for registering an admin
type AdminRegistrationRequest struct {
	Email           string `json:"email" binding:"required,email"`
	FirstName       string `json:"first_name" binding:"required"`
	LastName        string `json:"last_name" binding:"required"`
	Password        string `json:"password" binding:"required,min=6"`
	PasswordConfirm string `json:"password_confirm" binding:"required,min=6"`
	SecretKey       string `json:"secret_key" binding:"required"`
}

// CustomerRegistrationRequest represents the expected request body for registering a customer
type CustomerRegistrationRequest struct {
	Email           string `json:"email" binding:"required,email"`
	FirstName       string `json:"first_name" binding:"required"`
	LastName        string `json:"last_name" binding:"required"`
	Password        string `json:"password" binding:"required,min=6"`
	PasswordConfirm string `json:"password_confirm" binding:"required,min=6"`
}

// RegistrationSuccessResponse represents the response body for a successful registration
type RegistrationSuccessResponse struct {
	Msg          string      `json:"msg" example:"Account registered successfully"`
	User         models.User `json:"user"`
	AccessToken  string      `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string      `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// ErrorResponse represents the response body for an error
type ErrorResponse struct {
	Error string `json:"error" example:"Validation failed"`
}
