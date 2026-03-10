package main

import (
	"flathex/internal/adapters/memory"
	"flathex/internal/core/notifications"
	"flathex/internal/core/projects"
	"flathex/internal/core/tasks"
	"flathex/internal/platform/config"
	"flathex/internal/presentation"
	"log/slog"
	"os"
)

func main() {
	cfg := config.Load()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// ── Adapters (outbound) ───────────────────────────────────────────────────
	// Swap memory.* for postgres.* / smtp.Mailer to go to production.
	clock := memory.RealClock{}
	taskRepo := memory.NewTaskRepository()
	projectRepo := memory.NewProjectRepository()
	sender := memory.NoOpSender{}

	// ── Core services ─────────────────────────────────────────────────────────
	taskSvc := tasks.NewService(taskRepo, clock)
	projectSvc := projects.NewService(projectRepo, clock)
	notifSvc := notifications.NewService(sender, clock)

	// ── Presentation (inbound) ────────────────────────────────────────────────
	taskHandler := presentation.NewTaskHandler(taskSvc, projectSvc, notifSvc)
	projectHandler := presentation.NewProjectHandler(projectSvc, taskSvc)
	e := presentation.NewRouter(taskHandler, projectHandler)

	slog.Info("TaskHex starting", "port", cfg.Port, "env", cfg.Environment)
	presentation.PrintRoutes(e)

	if err := e.Start(":" + cfg.Port); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
