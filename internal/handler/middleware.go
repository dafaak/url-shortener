package handler

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/dafaak/url-shortener/internal/models"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
			return
		}

		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {


			userCtx:=models.UserContext{
				Username:claims["username"].(string),
				Plan: claims["plan"].(string),
			}
			

			c.Set("user",userCtx)

			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		}
	}
}
