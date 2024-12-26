package db

import (
	"fmt"
	"log"
	"os"

	"github.com/cgzirim/ecommerce-api/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// OpenDbConnection opens a connection to the database
func OpenDbConnection() *gorm.DB {
	username := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	host := os.Getenv("DB_HOST")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", host, username, password, dbName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	DB = db

	return db
}

// MigrateDBSchemas creates database tables based on the models
func MigrateDBSchemas() {
	if DB == nil {
		log.Fatal("Database connection is not initialized. Call OpenDbConnection first.")
	}

	err := DB.AutoMigrate(
		&models.User{}, &models.Product{},
		&models.Order{}, &models.OrderItem{}, &models.Address{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database schemas: %v", err)
	}

	log.Println("Database schemas migrated successfully.")
}
