package configuration

import (
	"fmt"
	"os"
)

// Simple helper function to read an environment or return a default value
func GetEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

// Get environment variable or error if it doesn't exist
func GetRequiredEnv(key string) (string, error) {
	if value, exists := os.LookupEnv(key); exists {
		return value, nil
	}

	return "", fmt.Errorf("could not find required config key %v", key)
}
