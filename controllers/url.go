package controllers

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/haseeb/url-shortener/config"
	"github.com/haseeb/url-shortener/models"
	"github.com/haseeb/url-shortener/utils"
	"gorm.io/gorm"
)

const maxShortCodeAttempts = 10

// helper to get base URL
func getAppURL() string {
	base := os.Getenv("APP_URL")
	if base == "" {
		base = "http://localhost:8080"
	}
	return base
}

func isValidPublicHTTPURL(raw string) bool {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	return true
}

// ─── SHORTEN URL ─────────────────────────────────────────

func ShortenURL(c *gin.Context) {
	var body struct {
		OriginalURL string `json:"original_url"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := c.ShouldBindJSON(&body); err != nil || body.OriginalURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if !isValidPublicHTTPURL(body.OriginalURL) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "original_url must be a valid http(s) URL with a host"})
		return
	}

	userID := c.MustGet("userID").(uint)

	expiresIn := body.ExpiresIn
	if expiresIn == 0 {
		expiresIn = 7
	}

	original := strings.TrimSpace(body.OriginalURL)
	expiresAt := time.Now().Add(time.Duration(expiresIn) * 24 * time.Hour)

	var shortURL models.ShortURL
	for attempt := 0; attempt < maxShortCodeAttempts; attempt++ {
		shortCode := utils.GenerateShortCode(6)
		shortURL = models.ShortURL{
			UserID:      userID,
			OriginalURL: original,
			ShortCode:   shortCode,
			ExpiresAt:   expiresAt,
		}

		err := config.DB.Create(&shortURL).Error
		if err == nil {
			c.JSON(http.StatusCreated, gin.H{
				"message":      "URL shortened successfully!",
				"short_code":   shortCode,
				"short_url":    getAppURL() + "/" + shortCode,
				"original_url": original,
				"expires_at":   shortURL.ExpiresAt,
			})
			return
		}

		if !errors.Is(err, gorm.ErrDuplicatedKey) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not shorten URL"})
			return
		}
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate a unique short code"})
}

// ─── REDIRECT ────────────────────────────────────────────

func RedirectURL(c *gin.Context) {
	shortCode := c.Param("shortCode")

	var shortURL models.ShortURL
	if err := config.DB.Where("short_code = ?", shortCode).First(&shortURL).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	if time.Now().After(shortURL.ExpiresAt) {
		c.JSON(http.StatusGone, gin.H{"error": "This URL has expired"})
		return
	}

	go func() {
		click := models.Click{
			URLID:     shortURL.ID,
			IPAddress: c.ClientIP(),
			Device:    c.GetHeader("User-Agent"),
			ClickedAt: time.Now(),
		}
		if err := config.DB.Create(&click).Error; err != nil {
			log.Printf("click logging failed: %v", err)
		}
	}()

	c.Redirect(http.StatusFound, shortURL.OriginalURL)
}

// ─── LIST MY URLs ─────────────────────────────────────────

func GetMyURLs(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var urls []models.ShortURL
	config.DB.Where("user_id = ?", userID).Find(&urls)

	// Add full short_url to each URL
	type URLResponse struct {
		ID          uint      `json:"id"`
		UserID      uint      `json:"user_id"`
		OriginalURL string    `json:"original_url"`
		ShortCode   string    `json:"short_code"`
		ShortURL    string    `json:"short_url"`
		ExpiresAt   time.Time `json:"expires_at"`
		CreatedAt   time.Time `json:"created_at"`
	}

	var response []URLResponse
	for _, url := range urls {
		response = append(response, URLResponse{
			ID:          url.ID,
			UserID:      url.UserID,
			OriginalURL: url.OriginalURL,
			ShortCode:   url.ShortCode,
			ShortURL:    getAppURL() + "/" + url.ShortCode, // ✅ dynamic URL
			ExpiresAt:   url.ExpiresAt,
			CreatedAt:   url.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total": len(urls),
		"urls":  response,
	})
}

// ─── DELETE URL ──────────────────────────────────────────

func DeleteURL(c *gin.Context) {
	urlID := c.Param("id")
	userID := c.MustGet("userID").(uint)

	var shortURL models.ShortURL
	if err := config.DB.Where("id = ? AND user_id = ?", urlID, userID).First(&shortURL).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	config.DB.Delete(&shortURL)

	c.JSON(http.StatusOK, gin.H{"message": "URL deleted successfully!"})
}
