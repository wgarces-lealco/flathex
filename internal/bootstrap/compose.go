package bootstrap

import (
	"database/sql"
	"flathex/internal/adapters/memory"
	"flathex/internal/adapters/sqlite"
	"flathex/internal/core/notifications"
	"flathex/internal/core/projects"
	"flathex/internal/core/tasks"
	"flathex/internal/presentation"

	"github.com/labstack/echo/v4"
)

// Composition Root for the HTTP server: the only package that wires adapters,
// core services, and presentation together. Split into *_wiring.go files if this grows large.

// BuildEcho constructs Echo with all routes and middleware wired from cfg and db.
func BuildEcho(cfg Config, db *sql.DB) *echo.Echo {

	clock := memory.RealClock{}
	taskRepo := sqlite.NewTaskRepository(db)
	projectRepo := sqlite.NewProjectRepository(db)
	sender := memory.NoOpSender{}

	taskSvc := tasks.NewService(taskRepo, clock)
	projectSvc := projects.NewService(projectRepo, clock)
	notifSvc := notifications.NewService(sender, clock)

	taskHandler := presentation.NewTaskHandler(taskSvc, projectSvc, notifSvc)
	projectHandler := presentation.NewProjectHandler(projectSvc, taskSvc)

	return presentation.NewRouter(taskHandler, projectHandler, presentation.RouterConfig{
		RequestTimeout: cfg.RequestTimeout,
	})
}

// PrintRoutes delegates to presentation for route listing (keeps main free of presentation import).
func PrintRoutes(e *echo.Echo) {
	presentation.PrintRoutes(e)
}
