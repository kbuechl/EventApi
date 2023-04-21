package config

import (
	"fmt"
	"os"
)

type DBConfig struct {
	Database string
	User     string
	Password string
	Host     string
}

type OauthConfig struct {
	Secret      string
	Client      string
	RedirectUrl string
}

type Server struct {
	SessionCookieName string
	CookieSecret      string
}

type Cache struct {
	Address  string
	Password string
}

type Config struct {
	DB     DBConfig
	Oauth  OauthConfig
	Server Server
	Cache  Cache
}

// New returns a new Config struct
func New() *Config {
	return &Config{
		DB: DBConfig{
			User:     getRequiredEnv("POSTGRES_USER"),
			Password: getRequiredEnv("POSTGRES_PASSWORD"),
			Database: getEnv("POSTGRES_DB", "postgres"),
			Host:     getEnv("POSTGRES_HOST", "localhost"),
		},
		Oauth: OauthConfig{
			Secret:      getRequiredEnv("Secret"),
			Client:      getRequiredEnv("Client"),
			RedirectUrl: getEnv("redirect_url", ""),
		},
		Server: Server{
			SessionCookieName: "session",
			CookieSecret:      getRequiredEnv("COOKIE_SECRET"),
		},
		Cache: Cache{
			Address:  fmt.Sprintf("%v:6379", getEnv("REDIS_HOST", "redis")),
			Password: getEnv("REDIS_PASSWORD", ""),
		},
	}
}

// Simple helper function to read an environment or return a default value
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

// Get environment variable or panic if it doesn't exist
func getRequiredEnv(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	panic(fmt.Sprintf("Could not find required config key %v", key))
}
