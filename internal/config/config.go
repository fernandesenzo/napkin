package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv           string
	ServerPort       string
	RedisAddress     string
	RedisPassword    string
	AllowedOrigins   string
	RateLimitGetMax  int
	RateLimitPostMax int
	RateLimitWindow  time.Duration
	CodeLength       int
	MaxContentLength int
	DefaultTTL       time.Duration
}

const (
	defaultCodeLength       = 6
	defaultMaxContentLength = 400
	defaultTTL              = 24 * time.Hour
)

func Load() (Config, error) {
	cfg := Config{
		AppEnv:         os.Getenv("APP_ENV"),
		ServerPort:     os.Getenv("SERVER_PORT"),
		RedisAddress:   os.Getenv("REDIS_ADDR"),
		RedisPassword:  os.Getenv("REDIS_PASSWORD"),
		AllowedOrigins: os.Getenv("ALLOWED_ORIGINS"),
	}

	required := map[string]string{
		"SERVER_PORT":           cfg.ServerPort,
		"REDIS_ADDR":            cfg.RedisAddress,
		"ALLOWED_ORIGINS":       cfg.AllowedOrigins,
		"RATE_LIMIT_GET_MAX":    os.Getenv("RATE_LIMIT_GET_MAX"),
		"RATE_LIMIT_POST_MAX":   os.Getenv("RATE_LIMIT_POST_MAX"),
		"RATE_LIMIT_WINDOW_SEC": os.Getenv("RATE_LIMIT_WINDOW_SEC"),
	}
	for name, value := range required {
		if value == "" {
			return cfg, fmt.Errorf("config.Load: %s must be set", name)
		}
	}

	var err error
	if cfg.RateLimitGetMax, err = positiveInt("RATE_LIMIT_GET_MAX"); err != nil {
		return cfg, err
	}
	if cfg.RateLimitPostMax, err = positiveInt("RATE_LIMIT_POST_MAX"); err != nil {
		return cfg, err
	}
	windowSec, err := positiveInt("RATE_LIMIT_WINDOW_SEC")
	if err != nil {
		return cfg, err
	}
	cfg.RateLimitWindow = time.Duration(windowSec) * time.Second

	cfg.CodeLength = optionalInt("CODE_LENGTH", defaultCodeLength)
	cfg.MaxContentLength = optionalInt("MAX_CONTENT_LENGTH", defaultMaxContentLength)
	if cfg.DefaultTTL, err = optionalDuration("DEFAULT_TTL", defaultTTL); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func positiveInt(name string) (int, error) {
	raw := os.Getenv(name)
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("config.Load: %s must be a positive integer", name)
	}
	return n, nil
}

func optionalInt(name string, def int) int {
	raw := os.Getenv(name)
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func optionalDuration(name string, def time.Duration) (time.Duration, error) {
	raw := os.Getenv(name)
	if raw == "" {
		return def, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("config.Load: %s must be a valid duration (e.g. 24h): %w", name, err)
	}
	return d, nil
}
