package config

import (
	"fmt"
	"log"
	"os"

	"github.com/haseeb/url-shortener/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB is a global variable — any file can use this to talk to database
var DB *gorm.DB

func ConnectDB() {
	// Read values from .env file
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	// Connect to PostgreSQL
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}

	// Auto-create tables from our models
	db.AutoMigrate(
		&models.User{},
		&models.ShortURL{},
		&models.Click{},
	)

	DB = db
	fmt.Println("✅ Database connected successfully!")
}
