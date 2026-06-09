package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
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

	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET must be set")
	}

	config.ConnectDB()

	utils.StartCronJobs()

	r := gin.Default()

	// ─── CORS ─────────────────────────────────────────────
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",       // local frontend
			"https://your-app.vercel.app", // replace with your Vercel URL after deploy
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

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
