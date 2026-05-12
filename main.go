package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/haseeb/url-shortener/config"
	"github.com/haseeb/url-shortener/routes"
	"github.com/haseeb/url-shortener/utils"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("❌ Error loading .env file")
	}

	config.ConnectDB()

	// ✅ Start cron jobs in background
	utils.StartCronJobs()

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "URL Shortener is running 🚀",
		})
	})

	routes.SetupRoutes(r)

	port := os.Getenv("APP_PORT")
	r.Run(":" + port)
}
