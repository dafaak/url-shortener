package utils

import (
	"github.com/gin-gonic/gin"
)

func SendSuccess(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(code, gin.H{
		"message": message,
		"data":    data,
	})
}

func SendError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"message": message,
		"data":    nil, // Mantenemos la estructura pero con data nulo
	})
}
