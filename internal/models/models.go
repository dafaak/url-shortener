package models

import (
	"time"
)

// URL es la tabla principal
type URL struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	UserID         *string    `gorm:"index" json:"user_id,omitempty"`
	OriginalURL    string     `gorm:"not null" json:"original_url"`
	ShortCode      string     `gorm:"uniqueIndex;not null;size:15" json:"short_code"`
	CreatedAt      time.Time  `json:"created_at"`
	LastAccessedAt *time.Time `json:"last_accessed_at,omitempty"`
	ExpiresAt      *time.Time `gorm:"index" json:"expires_at,omitempty"`
	ClickCount     int        `gorm:"default:0" json:"click_count"`
}

// Metric es la tabla de analíticas
type Metric struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	URLID       uint      `gorm:"index" json:"url_id"`
	ClickedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"clicked_at"`
	IPAddress   string    `json:"ip_address"`
	CountryCode string    `json:"country_code"`
	Browser     string    `json:"browser"`
	OS          string    `json:"os"`
}

// Estructuras para las peticiones HTTP (DTOs)
type ShortenRequest struct {
	URL       string     `json:"url" binding:"required,url"`
	UserID    *string    `json:"user_id"`
	ExpiresAt *time.Time `json:"expires_at"`
}
