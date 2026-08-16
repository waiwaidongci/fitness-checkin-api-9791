package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port            string
	DBPath          string
	GinMode         string
	DefaultPageSize int
	MaxPageSize     int
}

func Load() Config {
	return Config{
		Port:            getenv("PORT", "18009"),
		DBPath:          getenv("DB_PATH", "fitness.db"),
		GinMode:         getenv("GIN_MODE", "release"),
		DefaultPageSize: getenvInt("DEFAULT_PAGE_SIZE", 10),
		MaxPageSize:     getenvInt("MAX_PAGE_SIZE", 100),
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
