package bootstrap

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port           string
	Environment    string
	SQLitePath     string
	RequestTimeout time.Duration
	SMTPHost       string
	SMTPPort       string
	SMTPFrom       string
}

func LoadConfig() Config {
	return Config{
		Port:           getEnv("PORT", "8080"),
		Environment:    getEnv("ENV", "development"),
		SQLitePath:     getEnv("SQLITE_PATH", "flathex.db"),
		RequestTimeout: getDuration("REQUEST_TIMEOUT_SECONDS", 30) * time.Second,
		SMTPHost:       getEnv("SMTP_HOST", "localhost"),
		SMTPPort:       getEnv("SMTP_PORT", "1025"),
		SMTPFrom:       getEnv("SMTP_FROM", "noreply@taskhex.dev"),
	}
}

func getDuration(key string, fallback int) time.Duration {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return time.Duration(n)
		}
	}
	return time.Duration(fallback)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
