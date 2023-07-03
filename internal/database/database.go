package database

import (
	"eventapi/internal/configuration"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func New(cfg *configuration.DBConfig) (*gorm.DB, error) {
	dsn := getConnectionString(cfg)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return db, nil
}

func getConnectionString(cfg *configuration.DBConfig) string {
	return fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=disable",
		cfg.Host,
		cfg.User,
		cfg.Password,
		cfg.Database,
		"5432")
}
