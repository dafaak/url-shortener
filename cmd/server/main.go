package main

import (
	"log"
	"os"

	"github.com/dafaak/url-shortener/internal/handler"
	"github.com/dafaak/url-shortener/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	pg, err := storage.NewPostgresStorage(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	rd, err := storage.NewRedisStorage(os.Getenv("REDIS_ADDR"))
	if err != nil {
		log.Fatal(err)
	}

	authH := &handler.AuthHandler{DB: pg}
	h := handler.NewURLHandler(pg, rd)
	r := gin.Default()

	// Auth público
	r.POST("/register", authH.Register)
	r.POST("/login", authH.Login)

	// Redirección pública
	r.GET("/:code", h.Redirect)

	// Lista de links pública

	// Ruta pública para la app de Link Bio
	r.GET("/public/users/:username/links", h.GetPublicLinks)
	// Rutas protegidas
	protected := r.Group("/api")
	protected.Use(handler.AuthMiddleware())
	{
		protected.POST("/shorten", h.Shorten)
		protected.GET("/stats/:code", h.GetStats)
		protected.GET("/user/:username/links", h.GetUserURLs)
		protected.PATCH("/links/:code/privacy", h.TogglePrivacy)
	}

	log.Println("Servidor corriendo en :8080")
	r.Run(":8080")
}
