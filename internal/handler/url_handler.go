package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dafaak/url-shortener/internal/models"
	"github.com/dafaak/url-shortener/internal/storage"
	"github.com/dafaak/url-shortener/utils"
	"github.com/gin-gonic/gin"
	"github.com/ua-parser/uap-go/uaparser"
	"gorm.io/gorm"
)

type URLHandler struct {
	DB    *storage.PostgresStorage
	Cache *storage.RedisStorage
}

func NewURLHandler(db *storage.PostgresStorage, cache *storage.RedisStorage) *URLHandler {
	return &URLHandler{DB: db, Cache: cache}
}

func (h *URLHandler) GetStats(c *gin.Context) {
	shortCode := c.Param("code")
	//userID, _ := c.Get("userID")
	userID := "dafaak"
	var urlObj models.URL

	// 1. Validar propiedad del link
	if err := h.DB.DB.Where("short_code = ? AND user_id = ?", shortCode, userID).First(&urlObj).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Enlace no encontrado"})
		return
	}

	stats := models.URLStats{
		TotalClicks:  urlObj.ClickCount,
		LastAccessed: urlObj.LastAccessedAt,
		Browsers:     make(map[string]int),
		OS:           make(map[string]int),
		Platforms:    make(map[string]int),
	}

	// 2. Ejecutar conteos agrupados directamente en la DB
	// Agrupar por Navegador
	h.aggregateMetric(urlObj.ID, "browser", stats.Browsers)

	// Agrupar por Sistema Operativo
	h.aggregateMetric(urlObj.ID, "os", stats.OS)

	// Agrupar por Plataforma (Mobile/Desktop)
	h.aggregateMetric(urlObj.ID, "platform", stats.Platforms)

	c.JSON(http.StatusOK, stats)
}

// Función auxiliar para evitar repetir código de agrupación
func (h *URLHandler) aggregateMetric(urlID uint, column string, targetMap map[string]int) {
	rows, err := h.DB.DB.Model(&models.Metric{}).
		Select(column+" as label, count(*) as total").
		Where("url_id = ?", urlID).
		Group(column).
		Rows()

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var label string
			var total int
			rows.Scan(&label, &total)
			if label == "" {
				label = "Unknown"
			}
			targetMap[label] = total
		}
	}
}

func (h *URLHandler) GetUserURLs(c *gin.Context) {
	userID := c.Param("userId")
	var urls []models.URL

	// Buscamos en la base de datos filtrando por user_id
	result := h.DB.DB.Where("user_id = ?", userID).Find(&urls)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener los enlaces"})
		return
	}

	c.JSON(http.StatusOK, urls)
}

// Shorten crea el link corto
func (h *URLHandler) Shorten(c *gin.Context) {
	var req models.ShortenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	var finalCode string
	fmt.Println("custom_code", req.CustomCode)
	if req.CustomCode != "" {

		// --- CASO ALIAS PERSONALIZADO ---
		// 1. Validar longitud o caracteres (opcional)
		if len(req.CustomCode) < 3 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "El alias debe tener al menos 3 caracteres"})
			return
		}

		// 2. Verificar si ya existe en la base de datos
		var exists int64
		h.DB.DB.Model(&models.URL{}).Where("short_code = ?", req.CustomCode).Count(&exists)
		if exists > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Este alias ya está en uso"})
			return
		}
		finalCode = req.CustomCode
	} else {
		// --- CASO ALEATORIO ---
		finalCode = utils.Encode(uint64(time.Now().UnixNano()))
	}

	urlObj := models.URL{
		OriginalURL: req.URL,
		ShortCode:   finalCode,
		UserID:      req.UserID,
		ExpiresAt:   req.ExpiresAt,
	}

	if err := h.DB.DB.Create(&urlObj).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo crear el link"})
		return
	}

	// Guardar en caché
	h.Cache.Cli.Set(c.Request.Context(), finalCode, urlObj.OriginalURL, 24*time.Hour)

	c.JSON(http.StatusCreated, gin.H{
		"short_url": "http://localhost:8080/" + finalCode,
		"code":      finalCode,
	})
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
	// 1. Obtenemos el ID de la URL
	var urlObj models.URL
	if err := h.DB.DB.Select("id").Where("short_code = ?", code).First(&urlObj).Error; err != nil {
		return
	}

	// 2. Parsear el User-Agent
	uaString := c.Request.UserAgent()
	parser := uaparser.NewFromSaved()
	client := parser.Parse(uaString)

	// 3. Crear el registro de métrica
	metric := models.Metric{
		URLID:     urlObj.ID,
		IPAddress: c.ClientIP(),
		Browser:   client.UserAgent.Family,
		OS:        client.Os.Family,
		Platform:  client.Device.Family,
		Referrer:  c.Request.Referer(),
	}

	// 4. Guardar en DB y actualizar contador global
	h.DB.DB.Create(&metric)
	h.DB.DB.Model(&urlObj).UpdateColumn("click_count", gorm.Expr("click_count + 1"))
	h.DB.DB.Model(&urlObj).UpdateColumn("last_accessed_at", time.Now())
}
