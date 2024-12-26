package main

import (
	"log"

	"github.com/cgzirim/ecommerce-api/controllers"
	"github.com/cgzirim/ecommerce-api/db"
	_ "github.com/cgzirim/ecommerce-api/docs"
	"github.com/cgzirim/ecommerce-api/middleware"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title E-Commerce API
// @version 1.0
// @description This is a sample e-commerce API documentation.
// @host localhost:8080
// @BasePath /v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	godotenv.Load()

	db.OpenDbConnection()
	db.MigrateDBSchemas()

	router := SetupRouter()

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	err := router.Run(":8080")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func SetupRouter() *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/v1")
	{
		v1.Use(middleware.LoadAuthUserMiddleware())

		// Auth routes
		v1.POST("/login", controllers.LoginUser)
		v1.POST("/register", controllers.RegisterCustomer)
		v1.POST("/register/admin", controllers.RegisterAdmin)

		// User routes
		v1.POST("/addresses", controllers.CreateAddress)
		v1.GET("/users/addresses", controllers.ListAddresses)

		// Product routes
		v1.GET("/products", controllers.ListProducts)
		v1.GET("/products/:id", controllers.GetProductByID)
		v1.POST("/products", controllers.CreateProduct)
		v1.PUT("/products/:id", controllers.UpdateProduct)
		v1.PATCH("/products/:id", controllers.PatchProduct)
		v1.DELETE("/products/:id", controllers.DeleteProduct)

		// Order routes
		v1.POST("/orders", controllers.CreateOrder)
		v1.GET("/orders/:user_id", controllers.ListOrders)
		v1.PATCH("/orders/:id/cancel", controllers.CancelOrder)
		v1.PATCH("/orders/:id/status", controllers.UpdateOrderStatus)

	}

	return r
}
