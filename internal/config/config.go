package config

import (
	"flag"
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration.
type Config struct {
	DefaultTimeout time.Duration
	Port           int
	MaxWorkers     int
	LogLevel       string
	Version        string
}

// Load loads configuration from environment variables and CLI flags.
func Load() *Config {
	cfg := &Config{Version: "1.0.0"}

	port := flag.Int("port", 8080, "HTTP server port")
	maxWorkers := flag.Int("workers", 100, "Maximum concurrent workers")
	timeout := flag.Duration("timeout", 10*time.Second, "Default request timeout")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")

	flag.Parse()

	cfg.Port = getEnvInt("PORT", *port)
	cfg.MaxWorkers = getEnvInt("MAX_WORKERS", *maxWorkers)
	cfg.DefaultTimeout = getEnvDuration("DEFAULT_TIMEOUT", *timeout)
	cfg.LogLevel = getEnvString("LOG_LEVEL", *logLevel)

	return cfg
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvString(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}
