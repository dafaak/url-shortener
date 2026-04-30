package handler

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/dafaak/url-shortener/internal/models"
	"github.com/dafaak/url-shortener/internal/storage"
	"github.com/dafaak/url-shortener/utils"
	"github.com/gin-gonic/gin"
	"github.com/ua-parser/uap-go/uaparser"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type URLHandler struct {
	DB     *storage.PostgresStorage
	Cache  *storage.RedisStorage
	Parser *uaparser.Parser
}

func NewURLHandler(db *storage.PostgresStorage, cache *storage.RedisStorage) *URLHandler {
	uaParser := uaparser.NewFromSaved()
	return &URLHandler{DB: db, Cache: cache, Parser: uaParser}
}

func (h *URLHandler) GetStats(c *gin.Context) {
	shortCode := c.Param("code")
	var urlObj models.URL

	userVal, existsCtx := c.Get("username")
	if !existsCtx {
		utils.SendError(c, http.StatusUnauthorized, "Sesión expirada o inválida")
		return
	}

	usernameStr := userVal.(string)

	if err := h.DB.DB.Where("short_code = ? AND username = ?", shortCode, usernameStr).First(&urlObj).Error; err != nil {
		utils.SendError(c, http.StatusNotFound, "No se encontró el enlace o no tienes permisos")
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

	utils.SendSuccess(c, http.StatusOK, "Estadísticas cargadas", stats)
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

func (h *URLHandler) GetPublicLinks(c *gin.Context) {
	username := c.Param("username")
	var urls []models.URL

	// Buscamos links del usuario que:
	// 1. Le pertenezcan
	// 2. No hayan expirado (o no tengan fecha de expiración)
	//now := time.Now()
	err := h.DB.DB.Where("username = ? AND is_public = ?", username, true).
		//Where("(expires_at IS NULL OR expires_at > ?)", time.Now()).
		Find(&urls).Error

	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Error al obtener links")
		fmt.Println("Error: ", err)
		return
	}

	// Mapeamos a nuestro struct limpio
	var response []models.PublicLink
	for _, u := range urls {
		response = append(response, models.PublicLink{
			ShortCode:   u.ShortCode,
			OriginalURL: u.OriginalURL,
		})
	}

	utils.SendSuccess(c, http.StatusOK, "Enlaces publicos obtenidos", response)

}

func (h *URLHandler) GetUserURLs(c *gin.Context) {
	// 1. Extraemos el usuario del contexto (inyectado por tu Middleware de JWT)
	user, exists := utils.GetUserFromContext(c)
	if !exists {
		utils.SendError(c, http.StatusUnauthorized, "Sesión no válida")
		return
	}

	var urls []models.URL

	// 2. Buscamos en la DB filtrando por el username que sacamos del TOKEN
	// Esto garantiza que un usuario solo vea SUS propios links.
	result := h.DB.DB.Where("username = ?", user.Username).Find(&urls)

	if result.Error != nil {
		utils.SendError(c, http.StatusInternalServerError, "Error al obtener los enlaces")
		return
	}

	// 3. Devolvemos los links
	utils.SendSuccess(c, http.StatusOK, "¡Links obtenidos!", urls)
	// c.JSON(http.StatusOK, urls)
}

// Shorten crea el link corto
func (h *URLHandler) Shorten(c *gin.Context) {
	var req models.ShortenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Datos inválidos")
		return
	}

	user, ok := utils.GetUserFromContext(c)

	if !ok {
		utils.SendError(c, http.StatusUnauthorized, "Usuario no identificado")
		return
	}

	// EXTRAER EL USERNAME DEL CONTEXTO (TOKEN)
	usernameStr := user.Username

	var finalCode string

	if req.CustomCode != "" && user.Plan == "premium" {

		// --- CASO ALIAS PERSONALIZADO ---
		// 1. Validar longitud o caracteres (opcional)
		if len(req.CustomCode) < 3 {
			utils.SendError(c, http.StatusBadRequest, "El alias debe tener al menos 3 caracteres")
			return
		}

		// 2. Verificar si ya existe en la base de datos
		var exists int64
		h.DB.DB.Model(&models.URL{}).Where("short_code = ?", req.CustomCode).Count(&exists)
		if exists > 0 {
			utils.SendError(c, http.StatusConflict, "Este alias ya está en uso")
			return
		}
		finalCode = req.CustomCode
	} else {
		// --- CASO ALEATORIO ---
		finalCode = utils.Encode(uint64(time.Now().UnixNano()))
	}

	isPublic := true
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}

	urlObj := models.URL{
		OriginalURL: req.URL,
		ShortCode:   finalCode,
		IsPublic:    &isPublic,
		Username:    &usernameStr,
		ExpiresAt:   req.ExpiresAt,
		Alias:       req.Alias,
	}

	if err := h.DB.DB.Create(&urlObj).Error; err != nil {
		utils.SendError(c, http.StatusInternalServerError, "No se pudo crear el link")
		return
	}

	h.Cache.Cli.Set(c.Request.Context(), finalCode, urlObj.OriginalURL, 24*time.Hour)

	utils.SendSuccess(c, http.StatusCreated, "¡Enlace creado correctamente!", urlObj)

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
		utils.SendError(c, http.StatusNotFound, "No encontrado")
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
	client := h.Parser.Parse(uaString)
	ip := c.ClientIP()

	dbPath := os.Getenv("GEOIP_DB_PATH")

	if dbPath == "" {
		dbPath = "./internal/storage/geoip/GeoLite2-Country.mmdb"
	}

	fmt.Printf("DEBUG: Buscando IP: %s en Path: %s\n", ip, dbPath)
	country := utils.GetCountryFromIP(ip, dbPath)
	fmt.Printf("DEBUG: Resultado obtenido: %s\n", country)

	// 3. Crear el registro de métrica
	metric := models.Metric{
		URLID:       urlObj.ID,
		IPAddress:   ip,
		CountryCode: country,
		Browser:     client.UserAgent.Family,
		OS:          client.Os.Family,
		Platform:    client.Device.Family,
		Referrer:    c.Request.Referer(),
	}

	// 4. Guardar en DB y actualizar contador global
	h.DB.DB.Create(&metric)
	h.DB.DB.Model(&urlObj).Updates(map[string]interface{}{
		"click_count":      gorm.Expr("click_count + 1"),
		"last_accessed_at": time.Now(),
	})
}

func (h *URLHandler) TogglePrivacy(c *gin.Context) {
	id := c.Param("id")

	userVal, _ := c.Get("username")
	usernameStr := userVal.(string)

	var urlObj models.URL

	result := h.DB.DB.Model(&urlObj).
		Clauses(clause.Returning{}).
		Where("id = ? AND username = ?", id, usernameStr).
		Update("is_public", gorm.Expr("NOT is_public")) // Delegamos la inversión a la DB

	if result.Error != nil {
		utils.SendError(c, http.StatusInternalServerError, "No se pudo actualizar la privacidad")
		return
	}

	if result.RowsAffected == 0 {
		utils.SendError(c, http.StatusNotFound, "Link no encontrado o no tienes permiso")
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Privacidad actulizada ", urlObj)
}

func (h *URLHandler) Delete(c *gin.Context) {
	linkID := c.Param("id")

	val, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sesión no válida"})
		return
	}

	userCtx, ok := val.(models.UserContext)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error de contexto de usuario"})
		return
	}

	var link models.URL
	if err := h.DB.DB.Where("id = ? AND username = ?", linkID, userCtx.Username).First(&link).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Enlace no encontrado o no tienes permiso"})
		return
	}

	if err := h.DB.DB.Delete(&link).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al eliminar"})
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Enlace eliminado correctamente", link)

}
