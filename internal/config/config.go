package config

import (
	"fmt"
	"os"
)

type Config struct {
	AppEnv        string
	Port          string
	DBPath        string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	LogLevel      string
	LogFormat     string
	RateLimit     int
	JWTSecret     string
}

func LoadConfig() *Config {
	return &Config{
		AppEnv:        getEnv("APP_ENV", "development"),
		Port:          getEnv("PORT", "8080"),
		DBPath:        getEnv("DB_PATH", "./data/minlink.db"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
		LogFormat:     getEnv("LOG_FORMAT", "json"),
		RateLimit:     getEnvInt("RATE_LIMIT", 1000),
		JWTSecret:     getEnv("JWT_SECRET", "minlink_jwt_secret_key_2026"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}
