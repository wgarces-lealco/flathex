package bootstrap

import (
	"log/slog"
	"os"
)

// InitLogger configures the global slog logger based on the active environment.
// Must be called before any other component logs.
func InitLogger(cfg Config) {
	level := slog.LevelInfo
	if cfg.Environment == "development" {
		level = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})))
}
