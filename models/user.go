package models

import "time"

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null"`
	Email     string    `json:"email" gorm:"unique;not null"`
	Password  string    `json:"-"` // json:"-" means never send password in response
	CreatedAt time.Time `json:"created_at"`

	// One user can have many URLs
	URLs []ShortURL `json:"urls,omitempty" gorm:"foreignKey:UserID"`
}
