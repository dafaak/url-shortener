package handler

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/dafaak/url-shortener/internal/models"
	"github.com/dafaak/url-shortener/internal/storage"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	user := models.User{Email: req.Email, Username: req.Username, Password: string(hashedPassword)}

	fmt.Println("request:", req.Email, req.Password, req.Username)

	if err := h.DB.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "El email o el username ya están en uso"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Usuario registrado"})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	var user models.User

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de petición inválido"})
		return
	}

	if err := h.DB.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Credenciales inválidas mail"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Credenciales inválidas pass"})
		return
	}

	// Generar JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.Username,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})

	t, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	c.JSON(http.StatusOK, gin.H{"token": t})
}
