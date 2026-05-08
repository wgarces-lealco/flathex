package main

import (
	"flathex/internal/bootstrap"
	"log/slog"
	"os"
)

func main() {
	cfg := bootstrap.LoadConfig()
	bootstrap.InitLogger(cfg)

	db, err := bootstrap.OpenDB(cfg)
	if err != nil {
		slog.Error("failed to open database", "path", cfg.SQLitePath, "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("database close", "error", err)
		}
	}()

	e := bootstrap.BuildEcho(cfg, db)

	slog.Info("TaskHex starting", "port", cfg.Port, "env", cfg.Environment, "db", cfg.SQLitePath)
	bootstrap.PrintRoutes(e)

	if err := e.Start(":" + cfg.Port); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
