package models

import "time"

type Click struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	URLID     uint      `json:"url_id"` // which short URL was clicked
	IPAddress string    `json:"ip_address"`
	Device    string    `json:"device"` // mobile / desktop
	ClickedAt time.Time `json:"clicked_at"`
}
