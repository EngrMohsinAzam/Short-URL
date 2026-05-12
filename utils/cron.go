package utils

import (
	"fmt"
	"time"

	"github.com/haseeb/url-shortener/config"
	"github.com/haseeb/url-shortener/models"
	"github.com/robfig/cron/v3"
)

func StartCronJobs() {
	// Create a new cron scheduler
	c := cron.New()

	// ─── Job 1: Delete expired URLs every 24 hours ───
	// "@daily" means run once every day at midnight
	// For testing we use "@every 1m" (every 1 minute)
	c.AddFunc("@every 1m", func() {
		deleteExpiredURLs()
	})

	// Start the scheduler in background
	c.Start()

	fmt.Println("✅ Cron jobs started!")
}

func deleteExpiredURLs() {
	now := time.Now()

	// Find all expired URLs
	var expiredURLs []models.ShortURL
	config.DB.Where("expires_at < ?", now).Find(&expiredURLs)

	if len(expiredURLs) == 0 {
		fmt.Println("🕐 Cron ran — No expired URLs found")
		return
	}

	// Delete their clicks first (to avoid foreign key errors)
	for _, url := range expiredURLs {
		config.DB.Where("url_id = ?", url.ID).Delete(&models.Click{})
	}

	// Now delete the expired URLs
	result := config.DB.Where("expires_at < ?", now).Delete(&models.ShortURL{})

	fmt.Printf("🗑️  Cron ran — Deleted %d expired URLs\n", result.RowsAffected)
}
