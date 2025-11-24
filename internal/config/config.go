package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port          string
	PostEndpoint  string
	BatchSize     int
	BatchInterval time.Duration
	RetryDelay    time.Duration
	RetryAttempts int
}

func LoadConfig() *Config {
	return &Config{
		Port:          getEnv("APP_PORT", "8080"),
		PostEndpoint:  getEnv("POST_ENDPOINT", "localhost"),
		BatchSize:     getEnvAsInt("BATCH_SIZE", 10),
		BatchInterval: time.Duration(getEnvAsInt("BATCH_INTERVAL_SECONDS", 300)) * time.Second,
		RetryDelay:    2 * time.Second,
		RetryAttempts: 3,
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valStr := getEnv(key, "")
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return defaultVal
}
