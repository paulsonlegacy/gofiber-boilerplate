package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func loadEnvFile(path string) {
	if err := godotenv.Load(path); err != nil {
		// Not fatal — in production, env vars are injected directly by the platform.
		_ = err
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}

func mustParseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		panic("config: invalid duration: " + s)
	}
	return d
}
