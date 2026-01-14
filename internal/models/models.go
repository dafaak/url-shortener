package models

import (
	"time"
)

// URL es la tabla principal
type URL struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	Username       *string    `gorm:"index" json:"username,omitempty"`
	OriginalURL    string     `gorm:"not null" json:"original_url"`
	ShortCode      string     `gorm:"uniqueIndex;not null;size:15" json:"short_code"`
	CreatedAt      time.Time  `json:"created_at"`
	LastAccessedAt *time.Time `json:"last_accessed_at,omitempty"`
	ExpiresAt      *time.Time `gorm:"index" json:"expires_at,omitempty"`
	ClickCount     int        `gorm:"default:0" json:"click_count"`
	Metrics        []Metric   `gorm:"foreignKey:URLID" json:"-"`
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
	Referrer    string    `gorm:"type:text" json:"referrer"`
	Platform    string    `gorm:"size:50" json:"platform"`
}

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Email    string `gorm:"uniqueIndex;not null" json:"email"`
	Username string `gorm:"uniqueIndex;not null" json:"username"`
	Password string `gorm:"not null" json:"-"` // "-" para que nunca viaje en el JSON
	URLs     []URL  `gorm:"foreignKey:Username;references:Username" json:"urls,omitempty"`
}

type URLStats struct {
	TotalClicks  int            `json:"total_clicks"`
	LastAccessed *time.Time     `json:"last_accessed"`
	Browsers     map[string]int `json:"browsers"`
	Platforms    map[string]int `json:"platforms"`
	OS           map[string]int `json:"os"`
}

// Estructuras para las peticiones HTTP (DTOs)
type ShortenRequest struct {
	URL        string     `json:"url" binding:"required,url"`
	CustomCode string     `json:"custom_code,omitempty"`
	Username   *string    `json:"username,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,min=3,max=20"` // Obligatorio y con límites
	Password string `json:"password" binding:"required,min=6"`
}
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
