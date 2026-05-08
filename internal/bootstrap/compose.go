package bootstrap

import (
	"flathex/internal/adapters/memory"
	"flathex/internal/core/notifications"
	"flathex/internal/core/projects"
	"flathex/internal/core/tasks"
	"flathex/internal/presentation"

	"github.com/labstack/echo/v4"
)

// Composition Root for the HTTP server: the only package that wires adapters,
// core services, and presentation together. Split into *_wiring.go files if this grows large.

// BuildEcho constructs Echo with all routes and middleware wired from cfg (for env-derived defaults later).
func BuildEcho(cfg Config) *echo.Echo {
	_ = cfg // wiring may read timeouts or feature flags from cfg later

	clock := memory.RealClock{}
	taskRepo := memory.NewTaskRepository()
	projectRepo := memory.NewProjectRepository()
	sender := memory.NoOpSender{}

	taskSvc := tasks.NewService(taskRepo, clock)
	projectSvc := projects.NewService(projectRepo, clock)
	notifSvc := notifications.NewService(sender, clock)

	taskHandler := presentation.NewTaskHandler(taskSvc, projectSvc, notifSvc)
	projectHandler := presentation.NewProjectHandler(projectSvc, taskSvc)

	return presentation.NewRouter(taskHandler, projectHandler)
}

// PrintRoutes delegates to presentation for route listing (keeps main free of presentation import).
func PrintRoutes(e *echo.Echo) {
	presentation.PrintRoutes(e)
}
