package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/haseeb/url-shortener/controllers"
	"github.com/haseeb/url-shortener/middleware"
)

func SetupRoutes(r *gin.Engine) {
	// ─── Apply rate limiter to ALL routes ───
	r.Use(middleware.RateLimiter())

	// ─── Auth routes (no token needed) ───
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)
	}

	// ─── URL routes (token required) ───
	urls := r.Group("/api/urls")
	urls.Use(middleware.AuthMiddleware())
	{
		urls.POST("/", controllers.ShortenURL)
		urls.GET("/", controllers.GetMyURLs)
		urls.DELETE("/:id", controllers.DeleteURL)
	}

	// ─── Analytics (token required) ───
	analytics := r.Group("/api/analytics")
	analytics.Use(middleware.AuthMiddleware())
	{
		analytics.GET("/:id", controllers.GetAnalytics)
	}

	// ─── Redirect (public) ───
	r.GET("/:shortCode", controllers.RedirectURL)
}
