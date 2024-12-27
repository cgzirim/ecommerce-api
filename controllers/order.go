package controllers

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/cgzirim/ecommerce-api/db"
	"github.com/cgzirim/ecommerce-api/dtos"
	"github.com/cgzirim/ecommerce-api/models"
	"github.com/gin-gonic/gin"
)

// CreateOrder godoc
// @Summary Create a new order
// @Description Allows a user to create a new order with the specified address and items.
// @Tags Order
// @Accept json
// @Produce json
// @Param input body dtos.CreateOrderRequest true "Order information"
// @Success 201 {object} models.Order "Order created successfully"
// @Failure 400 {object} dtos.ErrorResponse "Invalid input data"
// @Failure 401 {object} dtos.ErrorResponse "Unauthenticated, login is required"
// @Failure 500 {object} dtos.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /orders [post]
func CreateOrder(c *gin.Context) {
	var createOrderRequest dtos.CreateOrderRequest
	if err := c.ShouldBindJSON(&createOrderRequest); err != nil {
		handleValidationErrors(err, c)
		return
	}

	authUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{Error: "Unauthenticated, login is required"})
		return
	}

	user, ok := authUser.(models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{Error: "Unauthenticated, login is required"})
		return
	}

	// loop through the order items and validate the product ID and quantity, and
	// calculate the total order amount
	var orderTotal float64
	products := make(map[uint]models.Product)
	for _, item := range createOrderRequest.OrderItems {
		var product models.Product
		if err := db.DB.First(&product, item.ProductID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Invalid product ID: %d", item.ProductID),
			})
			return
		}

		if item.Quantity <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Quantity must be greater than 0 for product ID: %d", item.ProductID),
			})
			return
		}

		products[product.ID] = product
		orderTotal += product.Price * float64(item.Quantity)
	}

	order := models.Order{
		UserID:    user.ID,
		AddressID: createOrderRequest.AddressID,
		Total:     float64(orderTotal),
		Status:    models.OrderStatusPending,
	}

	if err := db.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// loop through the order items and create the order items
	for _, item := range createOrderRequest.OrderItems {
		product := products[item.ProductID]

		orderItem := models.OrderItem{
			OrderID:   order.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     product.Price * float64(item.Quantity),
		}

		if err := db.DB.Create(&orderItem).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

	}

	if err := db.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := db.DB.Preload("User").Preload("Address").Preload("OrderItems").First(&order, order.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}

// ListOrders godoc
// @Summary List orders for a specific user
// @Description Retrieve a paginated list of orders for a specific user. Non-admin users can only list their own orders. Admins can list orders for any user.
// @Tags Order
// @Accept json
// @Produce json
// @Param user_id path int true "User ID"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Number of orders per page" default(10)
// @Success 200 {object} dtos.OrderListResponse "Successfully retrieved the paginated list of orders"
// @Failure 400 {object} dtos.ErrorResponse "Invalid user ID or page parameters"
// @Failure 401 {object} dtos.ErrorResponse "Unauthenticated, login is required"
// @Failure 403 {object} dtos.ErrorResponse "Unauthorized, you can only view your own orders"
// @Failure 500 {object} dtos.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /orders/{user_id} [get]
func ListOrders(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil || userID <= 0 {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: "Invalid user ID"})
		return
	}

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

	authUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{Error: "Unauthenticated, login is required"})
		return
	}

	user := authUser.(models.User)

	if !user.IsAdmin() && user.ID != uint(userID) {
		c.JSON(http.StatusForbidden, dtos.ErrorResponse{Error: "Unauthorized, you can only view your own orders"})
		return
	}

	var orders []models.Order
	result := db.DB.Preload("User").Preload("Address").Preload("OrderItems").Where("user_id = ?", userID).Limit(pageSize).Offset(offset).Find(&orders)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: result.Error.Error()})
		return
	}

	var totalOrders int64
	db.DB.Model(&models.Order{}).Where("user_id = ?", userID).Count(&totalOrders)

	c.JSON(http.StatusOK, dtos.OrderListResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalOrders,
		TotalPages: int(math.Ceil(float64(totalOrders) / float64(pageSize))),
		Orders:     orders,
	})
}

// CancelOrder godoc
// @Summary Cancel an order
// @Description Allows the owner of an order to cancel it if it is still in the pending status.
// @Tags Order
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} models.Order "Order cancelled successfully"
// @Failure 400 {object} dtos.ErrorResponse "Invalid order ID or order cannot be cancelled"
// @Failure 401 {object} dtos.ErrorResponse "Unauthenticated, login is required"
// @Failure 403 {object} dtos.ErrorResponse "Unauthorized, you can only cancel your own orders"
// @Failure 404 {object} dtos.ErrorResponse "Order not found"
// @Failure 500 {object} dtos.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /orders/{id}/cancel [patch]
func CancelOrder(c *gin.Context) {
	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil || orderID <= 0 {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: "Invalid order ID"})
		return
	}

	authUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{Error: "Unauthenticated, login is required"})
		return
	}

	user := authUser.(models.User)

	var order models.Order
	result := db.DB.First(&order, orderID)
	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			c.JSON(http.StatusNotFound, dtos.ErrorResponse{Error: "Order not found"})
		} else {
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: result.Error.Error()})
		}
		return
	}

	if order.UserID != user.ID {
		c.JSON(http.StatusForbidden, dtos.ErrorResponse{Error: "Unauthorized, you can only cancel your own orders"})
		return
	}

	if order.Status != models.OrderStatusPending {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: "Order cannot be cancelled, it is not in pending state"})
		return
	}

	order.Status = models.OrderStatusCancelled
	if err := db.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}

// UpdateOrderStatus godoc
// @Summary Update the status of an order
// @Description Allows an admin to update the status of an order.
// @Tags Order
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Param input body dtos.UpdateOrderStatusRequest true "Order status update information"
// @Success 200 {object} models.Order "Order status updated successfully"
// @Failure 400 {object} dtos.ErrorResponse "Invalid order ID or status"
// @Failure 401 {object} dtos.ErrorResponse "Unauthenticated, login is required"
// @Failure 403 {object} dtos.ErrorResponse "Unauthorized, only admins can update order status"
// @Failure 404 {object} dtos.ErrorResponse "Order not found"
// @Failure 500 {object} dtos.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /orders/{id}/status [patch]
func UpdateOrderStatus(c *gin.Context) {
	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil || orderID <= 0 {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: "Invalid order ID"})
		return
	}

	authUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{Error: "Unauthenticated, login is required"})
		return
	}

	user := authUser.(models.User)

	if !user.IsAdmin() {
		c.JSON(http.StatusForbidden, dtos.ErrorResponse{Error: "Unauthorized, only admins can update order status"})
		return
	}

	var order models.Order
	result := db.DB.First(&order, orderID)
	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			c.JSON(http.StatusNotFound, dtos.ErrorResponse{Error: "Order not found"})
		} else {
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: result.Error.Error()})
		}
		return
	}

	var req dtos.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleValidationErrors(err, c)
		return
	}

	if req.Status != models.OrderStatusPending && req.Status != models.OrderStatusCompleted && req.Status != models.OrderStatusCancelled {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: "Invalid order status"})
		return
	}

	order.Status = req.Status
	if err := db.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}
