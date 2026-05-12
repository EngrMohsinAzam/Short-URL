package models

import "time"

type ShortURL struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id"` // which user created this
	OriginalURL string    `json:"original_url" gorm:"not null"`
	ShortCode   string    `json:"short_code" gorm:"unique;not null"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`

	// One URL can have many clicks
	Clicks []Click `json:"clicks,omitempty" gorm:"foreignKey:URLID"`
}
