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

	// 1. Inicializar Almacenamiento
	pg, err := storage.NewPostgresStorage(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	rd, err := storage.NewRedisStorage(os.Getenv("REDIS_ADDR"))
	if err != nil {
		log.Fatal(err)
	}

	// 2. Handler y Rutas
	authH := &handler.AuthHandler{DB: pg}
	h := handler.NewURLHandler(pg, rd)
	r := gin.Default()

	// Auth público
	r.POST("/register", authH.Register)
	r.POST("/login", authH.Login)

	// Redirección pública
	r.GET("/:code", h.Redirect)
	//r.GET("/:user/:code", h.Redirect)

	// Lista de links pública
	r.GET("/user/:username/links", h.GetUserURLs)

	//shorten público
	r.POST("/shorten", h.Shorten)

	// Rutas protegidas
	api := r.Group("/api")
	api.Use(handler.AuthMiddleware())
	{
		r.GET("/stats/:code", h.GetStats)
	}

	log.Println("Servidor corriendo en :8080")
	r.Run(":8080")
}
