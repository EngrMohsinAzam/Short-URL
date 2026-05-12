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
	// Load .env if it exists (local dev)
	// On Railway env variables are injected automatically
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ No .env file found — using environment variables")
	}

	config.ConnectDB()

	utils.StartCronJobs()

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "URL Shortener is running 🚀",
		})
	})

	routes.SetupRoutes(r)

	// Use Railway's PORT variable or fallback to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("APP_PORT")
	}
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
