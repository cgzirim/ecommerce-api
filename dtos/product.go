package dtos

import "github.com/cgzirim/ecommerce-api/models"

// ProductListResponse represents the response body for a successful product listing
type ProductListResponse struct {
	Page       int              `json:"page" example:"2"`
	PageSize   int              `json:"page_size" example:"10"`
	TotalCount int64            `json:"total_count" example:"100"`
	TotalPages int64            `json:"total_pages" example:"10"`
	Products   []models.Product `json:"products"`
}

// CreateProductRequest represents the expected request body for creating a product
type CreateProductRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description" binding:"required"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	Stock       int     `json:"stock" binding:"required,gt=-1"`
	Category    string  `json:"category" binding:"required"`
}

// PatchProductRequest represents the expected request body for updating a product
type PatchProductRequest struct {
	Name        string  `json:"name" binding:"omitempty"`
	Description string  `json:"description" binding:"omitempty"`
	Price       float64 `json:"price" binding:"omitempty,gt=0"`
	Stock       int     `json:"stock" binding:"omitempty,gt=-1"`
	Category    string  `json:"category" binding:"omitempty"`
}
