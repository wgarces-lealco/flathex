package presentation

import (
	"fmt"
	"sort"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// NewRouter registers all routes and returns the configured Echo instance.
func NewRouter(tasks *TaskHandler, projects *ProjectHandler) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	// Tasks
	e.POST("/tasks", tasks.Create)
	e.GET("/tasks", tasks.ListAll)
	e.GET("/tasks/:id", tasks.GetByID)
	e.DELETE("/tasks/:id", tasks.Delete)
	e.PATCH("/tasks/:id/start", tasks.Start)
	e.PATCH("/tasks/:id/complete", tasks.Complete)
	e.PATCH("/tasks/:id/reopen", tasks.Reopen)

	// Projects
	e.POST("/projects", projects.Create)
	e.GET("/projects", projects.ListAll)
	e.GET("/projects/:id", projects.GetByID)
	e.GET("/projects/:id/tasks", projects.GetWithTasks)
	e.PATCH("/projects/:id/rename", projects.Rename)
	e.DELETE("/projects/:id", projects.Delete)

	return e
}

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
