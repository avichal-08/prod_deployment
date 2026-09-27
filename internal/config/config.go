package config

import (
	"os"
	"time"
)

type Config struct {
	Port            string
	Environment     string
	Version         string
	Hostname        string
	ShutdownTimeout time.Duration
}

func Load() Config {
	hostname := getEnv("HOSTNAME", "")
	if hostname == "" {
		if h, err := os.Hostname(); err == nil {
			hostname = h
		} else {
			hostname = "unknown"
		}
	}

	shutdownTimeout := 10 * time.Second
	if val := os.Getenv("SHUTDOWN_TIMEOUT"); val != "" {
		if parsed, err := time.ParseDuration(val); err == nil {
			shutdownTimeout = parsed
		}
	}

	return Config{
		Port:            getEnv("PORT", "8080"),
		Environment:     getEnv("APP_ENV", "development"),
		Version:         getEnv("APP_VERSION", "0.1.0"),
		Hostname:        hostname,
		ShutdownTimeout: shutdownTimeout,
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}
