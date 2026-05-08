package main

import (
	"flathex/internal/bootstrap"
	"log/slog"
	"os"
)

func main() {
	cfg := bootstrap.LoadConfig()
	bootstrap.InitLogger(cfg)

	e := bootstrap.BuildEcho(cfg)

	slog.Info("TaskHex starting", "port", cfg.Port, "env", cfg.Environment)
	bootstrap.PrintRoutes(e)

	if err := e.Start(":" + cfg.Port); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
