package config

import (
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from .env file
func LoadEnv() error {
	return godotenv.Load()
}

// GetEnv retrieves an environment variable with a fallback value
func GetEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// GetDatabaseConfig returns database configuration from environment variables
func GetDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     GetEnv("DB_HOST", "localhost"),
		Port:     GetEnv("DB_PORT", "5432"),
		User:     GetEnv("DB_USER", "postgres"),
		Password: GetEnv("DB_PASSWORD", "password"),
		Name:     GetEnv("DB_NAME", "tribute_db"),
		SSLMode:  GetEnv("DB_SSL_MODE", "disable"),
	}
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// GetRedisConfig returns Redis configuration from environment variables
func GetRedisConfig() RedisConfig {
	return RedisConfig{
		Host:     GetEnv("REDIS_HOST", "localhost"),
		Port:     GetEnv("REDIS_PORT", "6379"),
		Password: GetEnv("REDIS_PASSWORD", ""),
		DB:       0, // Default to DB 0
	}
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret string
	Expiry string
}

// GetJWTConfig returns JWT configuration from environment variables
func GetJWTConfig() JWTConfig {
	return JWTConfig{
		Secret: GetEnv("JWT_SECRET", "default-secret-key"),
		Expiry: GetEnv("JWT_EXPIRY", "24h"),
	}
}
