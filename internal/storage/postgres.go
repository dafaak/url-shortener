package storage

import (
	"fmt"

	"github.com/dafaak/url-shortener/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresStorage struct {
	DB *gorm.DB
}

func NewPostgresStorage(dsn string) (*PostgresStorage, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("fallo al conectar a postgres: %w", err)
	}

	// Esto crea las tablas automáticamente si no existen
	err = db.AutoMigrate(&models.User{}, &models.URL{}, &models.Metric{})
	if err != nil {
		return nil, fmt.Errorf("error migrando tablas: %w", err)
	}

	return &PostgresStorage{DB: db}, nil
}
