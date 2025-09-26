package database

import (
	"fmt"

	"life-certificates/internal/domain"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// New initialises a gorm DB connection using PostgreSQL with the provided DSN.
func New(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	return db, nil
}

// Migrate applies the schema required for the service.
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&domain.Participant{}, &domain.LifeCertificate{}, &domain.FRIdentity{}, &domain.Member{}); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}
	return nil
}
