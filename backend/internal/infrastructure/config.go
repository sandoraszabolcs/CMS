package infrastructure

import (
	"fmt"
	"os"
	"time"
)

// Config holds all application configuration from environment variables.
type Config struct {
	DBURL             string
	RedisAddr         string
	HTTPPort          string
	LogLevel          string
	SimulatorInterval time.Duration
	ODRefreshInterval time.Duration
}

// LoadConfig reads configuration from environment variables with defaults.
func LoadConfig() (Config, error) {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		return Config{}, fmt.Errorf("DB_URL environment variable is required")
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		return Config{}, fmt.Errorf("REDIS_ADDR environment variable is required")
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	simInterval, err := parseDurationEnv("SIMULATOR_INTERVAL", 2*time.Second)
	if err != nil {
		return Config{}, fmt.Errorf("invalid SIMULATOR_INTERVAL: %w", err)
	}

	odInterval, err := parseDurationEnv("OD_REFRESH_INTERVAL", 30*time.Second)
	if err != nil {
		return Config{}, fmt.Errorf("invalid OD_REFRESH_INTERVAL: %w", err)
	}

	return Config{
		DBURL:             dbURL,
		RedisAddr:         redisAddr,
		HTTPPort:          httpPort,
		LogLevel:          logLevel,
		SimulatorInterval: simInterval,
		ODRefreshInterval: odInterval,
	}, nil
}

func parseDurationEnv(key string, defaultVal time.Duration) (time.Duration, error) {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal, nil
	}
	return time.ParseDuration(val)
}
