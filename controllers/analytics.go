package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/haseeb/url-shortener/config"
	"github.com/haseeb/url-shortener/models"
)

// Response structure for analytics
type ClickPerDay struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

func GetAnalytics(c *gin.Context) {
	// 1. Get URL id from params and logged in user
	urlID := c.Param("id")
	userID := c.MustGet("userID").(uint)

	// 2. Make sure this URL belongs to the logged in user
	var shortURL models.ShortURL
	if err := config.DB.Where("id = ? AND user_id = ?", urlID, userID).First(&shortURL).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	// 3. Count total clicks
	var totalClicks int64
	config.DB.Model(&models.Click{}).Where("url_id = ?", urlID).Count(&totalClicks)

	// 4. Get clicks grouped by date
	var clicksPerDay []ClickPerDay
	config.DB.Model(&models.Click{}).
		Select("DATE(clicked_at) as date, COUNT(*) as count").
		Where("url_id = ?", urlID).
		Group("DATE(clicked_at)").
		Order("date ASC").
		Scan(&clicksPerDay)

	// 5. Get last 5 clicks detail
	var recentClicks []models.Click
	config.DB.Where("url_id = ?", urlID).
		Order("clicked_at DESC").
		Limit(5).
		Find(&recentClicks)

	// 6. Return everything
	c.JSON(http.StatusOK, gin.H{
		"url": gin.H{
			"id":           shortURL.ID,
			"short_code":   shortURL.ShortCode,
			"original_url": shortURL.OriginalURL,
			"short_url":    getAppURL() + "/" + shortURL.ShortCode,
			"expires_at":   shortURL.ExpiresAt,
		},
		"analytics": gin.H{
			"total_clicks":   totalClicks,
			"clicks_per_day": clicksPerDay,
			"recent_clicks":  recentClicks,
		},
	})
}
