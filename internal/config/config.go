package config

import (
	"os"
	"time"
)

type Config struct {
	Port string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	HTTPReadTimeout   time.Duration
	HTTPWriteTimeout   time.Duration
	HTTPIdleTimeout    time.Duration

	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
}

func Load() Config {
	return Config{
		Port: getEnv("PORT", "8080"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "sofreserve"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		HTTPReadTimeout:  parseDuration("HTTP_READ_TIMEOUT", 5*time.Second),
		HTTPWriteTimeout: parseDuration("HTTP_WRITE_TIMEOUT", 10*time.Second),
		HTTPIdleTimeout:  parseDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),

		DBMaxOpenConns:    20,
		DBMaxIdleConns:    10,
		DBConnMaxLifetime: 30 * time.Minute,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
	}
	return fallback
}