package bootstrap

import "os"

type Config struct {
	Port        string
	Environment string
	SQLitePath  string
	SMTPHost    string
	SMTPPort    string
	SMTPFrom    string
}

func LoadConfig() Config {
	return Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENV", "development"),
		SQLitePath:  getEnv("SQLITE_PATH", "flathex.db"),
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
