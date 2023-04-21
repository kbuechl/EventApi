package database

import (
	"eventapi/config"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var instance *gorm.DB

func init() {
	dsn := getConnectionString()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		println("Cannot connect to postgres")
		panic(err)
	}

	instance = db
}
func getConnectionString() string {
	config := config.New()

	return fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=disable",
		config.DB.Host,
		config.DB.User,
		config.DB.Password,
		config.DB.Database,
		"5432")
}

func NewClient() *gorm.DB {
	return instance
}
