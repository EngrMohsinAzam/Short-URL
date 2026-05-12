package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/haseeb/url-shortener/config"
	"github.com/haseeb/url-shortener/models"
	"github.com/haseeb/url-shortener/utils"
)

// ─── SHORTEN URL ─────────────────────────────────────────

func ShortenURL(c *gin.Context) {
	// 1. Read request body
	var body struct {
		OriginalURL string `json:"original_url"`
		ExpiresIn   int    `json:"expires_in"` // days until expiry
	}

	if err := c.ShouldBindJSON(&body); err != nil || body.OriginalURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// 2. Get logged in user ID from middleware
	userID := c.MustGet("userID").(uint)

	// 3. Generate unique short code
	shortCode := utils.GenerateShortCode(6)

	// 4. Set expiry (default 7 days if not provided)
	expiresIn := body.ExpiresIn
	if expiresIn == 0 {
		expiresIn = 7
	}

	// 5. Save to DB
	shortURL := models.ShortURL{
		UserID:      userID,
		OriginalURL: body.OriginalURL,
		ShortCode:   shortCode,
		ExpiresAt:   time.Now().Add(time.Duration(expiresIn) * 24 * time.Hour),
	}

	if err := config.DB.Create(&shortURL).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not shorten URL"})
		return
	}

	// 6. Return the short URL
	c.JSON(http.StatusCreated, gin.H{
		"message":      "URL shortened successfully!",
		"short_code":   shortCode,
		"short_url":    "http://localhost:8080/" + shortCode,
		"original_url": body.OriginalURL,
		"expires_at":   shortURL.ExpiresAt,
	})
}

// ─── REDIRECT ────────────────────────────────────────────

func RedirectURL(c *gin.Context) {
	// 1. Get short code from URL param
	shortCode := c.Param("shortCode")

	// 2. Find it in DB
	var shortURL models.ShortURL
	if err := config.DB.Where("short_code = ?", shortCode).First(&shortURL).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	// 3. Check if expired
	if time.Now().After(shortURL.ExpiresAt) {
		c.JSON(http.StatusGone, gin.H{"error": "This URL has expired"})
		return
	}

	// 4. Track click in background using Goroutine (does not slow down redirect)
	go func() {
		click := models.Click{
			URLID:     shortURL.ID,
			IPAddress: c.ClientIP(),
			Device:    c.GetHeader("User-Agent"),
			ClickedAt: time.Now(),
		}
		config.DB.Create(&click)
	}()

	// 5. Redirect user to original URL
	c.Redirect(http.StatusMovedPermanently, shortURL.OriginalURL)
}

// ─── LIST MY URLs ─────────────────────────────────────────

func GetMyURLs(c *gin.Context) {
	// 1. Get logged in user ID
	userID := c.MustGet("userID").(uint)

	// 2. Get all URLs for this user
	var urls []models.ShortURL
	config.DB.Where("user_id = ?", userID).Find(&urls)

	// 3. Return them
	c.JSON(http.StatusOK, gin.H{
		"total": len(urls),
		"urls":  urls,
	})
}

// ─── DELETE URL ──────────────────────────────────────────

func DeleteURL(c *gin.Context) {
	// 1. Get URL id from params and logged in user
	urlID := c.Param("id")
	userID := c.MustGet("userID").(uint)

	// 2. Find the URL and make sure it belongs to this user
	var shortURL models.ShortURL
	if err := config.DB.Where("id = ? AND user_id = ?", urlID, userID).First(&shortURL).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	// 3. Delete it
	config.DB.Delete(&shortURL)

	c.JSON(http.StatusOK, gin.H{"message": "URL deleted successfully!"})
}
