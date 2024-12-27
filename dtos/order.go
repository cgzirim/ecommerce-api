package dtos

import (
	"github.com/cgzirim/ecommerce-api/models"
)

// OrderItemRequest represents the request to add an item to an order
type OrderItemRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,gt=0"`
}

// CreateOrderRequest represents the expected request body for creating an order
type CreateOrderRequest struct {
	AddressID  uint               `json:"address_id" binding:"required"`
	OrderItems []OrderItemRequest `json:"order_items" binding:"required,min=1"`
}

// OrderDetail represents the response body for a successful order creation
type OrderListResponse struct {
	Page       int            `json:"page" example:"1"`
	PageSize   int            `json:"page_size" example:"10"`
	TotalCount int64          `json:"total_count" example:"100"`
	TotalPages int            `json:"total_pages" example:"10"`
	Orders     []models.Order `json:"orders"`
}

// OrderDetail represents the response body for a successful order creation
type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required"`
}
