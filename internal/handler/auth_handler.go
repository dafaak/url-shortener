package handler

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dafaak/url-shortener/internal/models"
	"github.com/dafaak/url-shortener/internal/storage"
	"github.com/dafaak/url-shortener/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	DB *storage.PostgresStorage
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Formato de datos inválido")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Error al procesar la seguridad de la cuenta")
		return
	}

	user := models.User{
		Email:    req.Email,
		Username: req.Username,
		Password: string(hashedPassword),
		Plan:     "free",
	}

	if err := h.DB.DB.Create(&user).Error; err != nil {
		utils.SendError(c, http.StatusConflict, "El email o el nombre de usuario ya están registrados")
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "¡Cuenta creada con éxito! Ya puedes iniciar sesión.", nil)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	var user models.User

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Formato de petición inválido")
		return
	}

	if err := h.DB.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		utils.SendError(c, http.StatusUnauthorized, "Credenciales inválidas")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		utils.SendError(c, http.StatusUnauthorized, "Credenciales inválidas")
		return
	}

	var currentCount int64
	h.DB.DB.Model(&models.URL{}).Where("username = ?", user.Username).Count(&currentCount)

	freeLimitStr := os.Getenv("FREE_LINKS_LIMIT")
	limit, err := strconv.Atoi(freeLimitStr)
	if err != nil {
		limit = 10
	}

	if user.Plan == "premium" {
		limit = 1000
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"plan":     user.Plan,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	})

	t, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Error al generar el acceso")
		return
	}

	response := models.LoginResponse{
		Token:        t,
		Username:     user.Username,
		Plan:         user.Plan,
		Limit:        limit,
		CurrentLinks: int(currentCount),
	}

	utils.SendSuccess(c, http.StatusOK, "¡Bienvenido de nuevo!", response)
}
