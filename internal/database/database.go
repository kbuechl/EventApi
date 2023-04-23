package database

import (
	"eventapi/internal/configuration"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DBConfig struct {
	Database string
	User     string
	Password string
	Host     string
}

func New() *gorm.DB {
	dsn := getConnectionString()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		println("Cannot connect to postgres")
		panic(err)
	}

	return db
}

func getConnectionString() string {
	c := configure()

	return fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=disable",
		c.Host,
		c.User,
		c.Password,
		c.Database,
		"5432")
}

func configure() *DBConfig {
	u, err := configuration.GetRequiredEnv("POSTGRES_USER")

	if err != nil {
		panic(err)
	}

	p, err := configuration.GetRequiredEnv("POSTGRES_PASSWORD")

	if err != nil {
		panic(err)
	}

	return &DBConfig{
		User:     u,
		Password: p,
		Database: configuration.GetEnv("POSTGRES_DB", "postgres"),
		Host:     configuration.GetEnv("POSTGRES_HOST", "localhost"),
	}
}
