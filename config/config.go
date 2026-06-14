package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	HTTPAddr           string
	DBPath             string
	ShutdownTimeout    time.Duration
	WorkerPollInterval time.Duration
	JWTSecret          string
}

func Load() (Config, error) {
	workerPollInterval, err := parseDuration("WORKER_POLL_INTERVAL", 10*time.Second)
	if err != nil {
		return Config{}, err
	}

	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET must not be empty")
	}

	cfg := Config{
		HTTPAddr:           getEnv("HTTP_ADDR", ":8080"),
		DBPath:             getEnv("DB_PATH", "reminder.db"),
		ShutdownTimeout:    10 * time.Second,
		WorkerPollInterval: workerPollInterval,
		JWTSecret:          jwtSecret,
	}

	if cfg.DBPath == "" {
		return Config{}, fmt.Errorf("DB_PATH must not be empty")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseDuration(key string, fallback time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("%s: invalid duration: %w", key, err)
	}
	return d, nil
}
