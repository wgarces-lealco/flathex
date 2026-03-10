package config

import "os"

type Config struct {
	Port        string
	Environment string
	SMTPHost    string
	SMTPPort    string
	SMTPFrom    string
}

func Load() Config {
	return Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENV", "development"),
		SMTPHost:    getEnv("SMTP_HOST", "localhost"),
		SMTPPort:    getEnv("SMTP_PORT", "1025"),
		SMTPFrom:    getEnv("SMTP_FROM", "noreply@taskhex.dev"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
