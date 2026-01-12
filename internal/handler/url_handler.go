package handler

import (
	"net/http"
	"time"

	"github.com/dafaak/url-shortener/internal/models"
	"github.com/dafaak/url-shortener/internal/storage"
	"github.com/dafaak/url-shortener/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type URLHandler struct {
	DB    *storage.PostgresStorage
	Cache *storage.RedisStorage
}

func NewURLHandler(db *storage.PostgresStorage, cache *storage.RedisStorage) *URLHandler {
	return &URLHandler{DB: db, Cache: cache}
}

// Shorten crea el link corto
func (h *URLHandler) Shorten(c *gin.Context) {
	var req models.ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL inválida"})
		return
	}

	code := utils.Encode(uint64(time.Now().UnixNano()))
	urlObj := models.URL{
		OriginalURL: req.URL,
		ShortCode:   code,
		UserID:      req.UserID,
		ExpiresAt:   req.ExpiresAt,
	}

	if err := h.DB.DB.Create(&urlObj).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al guardar"})
		return
	}

	// Guardar en caché por 24h
	h.Cache.Cli.Set(c.Request.Context(), code, urlObj.OriginalURL, 24*time.Hour)

	c.JSON(http.StatusCreated, gin.H{"short_url": "http://localhost:8080/" + code})
}

// Redirect busca y redirige
func (h *URLHandler) Redirect(c *gin.Context) {
	code := c.Param("code")
	ctx := c.Request.Context()

	// 1. Buscar en Redis
	val, err := h.Cache.Cli.Get(ctx, code).Result()
	if err == nil {
		go h.recordMetric(code, c) // Goroutine para métricas
		c.Redirect(http.StatusMovedPermanently, val)
		return
	}

	// 2. Buscar en Postgres
	var urlObj models.URL
	if err := h.DB.DB.Where("short_code = ?", code).First(&urlObj).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No encontrado"})
		return
	}

	// 3. Actualizar caché y métricas
	h.Cache.Cli.Set(ctx, code, urlObj.OriginalURL, 24*time.Hour)
	go h.recordMetric(code, c)

	c.Redirect(http.StatusMovedPermanently, urlObj.OriginalURL)
}

func (h *URLHandler) recordMetric(code string, c *gin.Context) {
	// Aquí guardarías en models.Metric usando h.DB.DB
	// Por ahora solo actualizamos el contador global
	h.DB.DB.Model(&models.URL{}).Where("short_code = ?", code).
		UpdateColumn("click_count", gorm.Expr("click_count + 1")).
		UpdateColumn("last_accessed_at", time.Now())
}
