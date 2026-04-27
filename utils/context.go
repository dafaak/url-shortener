package utils

import (
	"github.com/dafaak/url-shortener/internal/models"
	"github.com/gin-gonic/gin"
)

func GetUserFromContext(c *gin.Context) (models.UserContext, bool) {
	user, exists := c.Get("user")
	if !exists {
		return models.UserContext{}, false
	}

	userVal, ok := user.(models.UserContext)
	return userVal, ok
}
