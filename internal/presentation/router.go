package presentation

import (
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// RouterConfig holds presentation-layer tunables supplied from bootstrap.
type RouterConfig struct {
	RequestTimeout time.Duration
}

// NewRouter registers all routes and middleware, returning the Echo instance.
func NewRouter(tasks *TaskHandler, projects *ProjectHandler, cfg RouterConfig) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// ── 4. Global JSON error handler ────────────────────────────────────────
	// Guarantees every error response (domain, panic, 404) is always JSON —
	// never the default Echo HTML page.
	e.HTTPErrorHandler = jsonErrorHandler

	// ── Middleware stack (order matters) ────────────────────────────────────
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	// ── 5. Per-request timeout ───────────────────────────────────────────────
	// Cancels the request context after cfg.RequestTimeout, preventing slow
	// queries from holding goroutines indefinitely. Uses Go's context mechanism
	// instead of the deprecated TimeoutWithConfig (which had data races).
	e.Use(middleware.ContextTimeoutWithConfig(middleware.ContextTimeoutConfig{
		Timeout: cfg.RequestTimeout,
	}))

	// ── 7. Request ID propagated into slog ──────────────────────────────────
	// Each handler's context carries a logger pre-seeded with request_id so
	// every downstream log line is correlated without manual plumbing.
	e.Use(requestIDLogger)

	// ── 3. Health check ─────────────────────────────────────────────────────
	// Minimal liveness probe — no auth, no business logic, always 200.
	e.GET("/healthz", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// Tasks
	e.POST("/tasks", tasks.Create)
	e.GET("/tasks", tasks.ListAll)
	e.GET("/tasks/:id", tasks.GetByID)
	e.DELETE("/tasks/:id", tasks.Delete)
	e.PATCH("/tasks/:id/start", tasks.Start)
	e.PATCH("/tasks/:id/complete", tasks.Complete)
	e.PATCH("/tasks/:id/reopen", tasks.Reopen)
	e.PATCH("/tasks/:id/cancel", tasks.Cancel)

	// Projects
	e.POST("/projects", projects.Create)
	e.GET("/projects", projects.ListAll)
	e.GET("/projects/:id", projects.GetByID)
	e.GET("/projects/:id/tasks", projects.GetWithTasks)
	e.PATCH("/projects/:id/rename", projects.Rename)
	e.DELETE("/projects/:id", projects.Delete)

	return e
}

// ── 4. jsonErrorHandler ───────────────────────────────────────────────────────
// Replaces Echo's default HTML error page with a consistent JSON envelope.
// Presentation layer owns the translation; the core never touches HTTP.
func jsonErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	msg := "internal server error"

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		if s, ok := he.Message.(string); ok {
			msg = s
		}
	}

	if !c.Response().Committed {
		if err := c.JSON(code, map[string]any{
			"error":      msg,
			"status":     code,
			"request_id": c.Response().Header().Get(echo.HeaderXRequestID),
		}); err != nil {
			slog.Error("failed to write error response", "error", err)
		}
	}
}

// ── 7. requestIDLogger ────────────────────────────────────────────────────────
// Middleware that stores a request-scoped logger in echo.Context so handlers
// can call loggerFrom(c) and get slog entries pre-tagged with request_id.
func requestIDLogger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		reqID := c.Response().Header().Get(echo.HeaderXRequestID)
		c.Set("logger", slog.With("request_id", reqID))
		return next(c)
	}
}

// LoggerFrom extracts the request-scoped slog logger stored by requestIDLogger.
// Falls back to the global logger if not set — safe to call from any handler.
func LoggerFrom(c echo.Context) *slog.Logger {
	if l, ok := c.Get("logger").(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

// ── PrintRoutes ───────────────────────────────────────────────────────────────

// PrintRoutes prints all registered routes to stdout, grouped by path prefix
// and sorted by (path, method) for easy scanning.
func PrintRoutes(e *echo.Echo) {
	type row struct{ method, path string }

	routes := e.Routes()
	rows := make([]row, 0, len(routes))
	for _, r := range routes {
		if r.Method == "echo_route_not_found" {
			continue
		}
		rows = append(rows, row{r.Method, r.Path})
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].path != rows[j].path {
			return rows[i].path < rows[j].path
		}
		return rows[i].method < rows[j].method
	})

	methodW, pathW := len("METHOD"), len("PATH")
	for _, r := range rows {
		if len(r.method) > methodW {
			methodW = len(r.method)
		}
		if len(r.path) > pathW {
			pathW = len(r.path)
		}
	}

	sep := fmt.Sprintf("├─%s─┬─%s─┤", strings.Repeat("─", methodW), strings.Repeat("─", pathW))
	top := fmt.Sprintf("┌─%s─┬─%s─┐", strings.Repeat("─", methodW), strings.Repeat("─", pathW))
	bot := fmt.Sprintf("└─%s─┴─%s─┘", strings.Repeat("─", methodW), strings.Repeat("─", pathW))

	fmt.Println(top)
	fmt.Printf("│ %-*s │ %-*s │\n", methodW, "METHOD", pathW, "PATH")

	prevPrefix := ""
	for _, r := range rows {
		prefix := routePrefix(r.path)
		if prefix != prevPrefix {
			fmt.Println(sep)
			prevPrefix = prefix
		}
		fmt.Printf("│ %-*s │ %-*s │\n", methodW, r.method, pathW, r.path)
	}

	fmt.Println(bot)
}

func routePrefix(path string) string {
	parts := strings.SplitN(strings.TrimPrefix(path, "/"), "/", 2)
	return "/" + parts[0]
}
