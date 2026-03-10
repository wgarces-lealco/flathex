package presentation

import (
	"context"
	"errors"
	"flathex/internal/core/projects"
	"flathex/internal/core/tasks"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ── Consumer-defined interfaces (Rule 1) ─────────────────────────────────────

type projectService interface {
	Create(ctx context.Context, id, name, description string) (*projects.Project, error)
	Rename(ctx context.Context, id, name string) (*projects.Project, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*projects.Project, error)
	ListAll(ctx context.Context) ([]*projects.Project, error)
}

type taskLister interface {
	ListByProject(ctx context.Context, projectID string) ([]*tasks.Task, error)
}

// ── DTOs ─────────────────────────────────────────────────────────────────────

type createProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type renameProjectRequest struct {
	Name string `json:"name"`
}

type projectResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	TaskCount   int    `json:"task_count"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type projectWithTasksResponse struct {
	projectResponse
	Tasks []taskResponse `json:"tasks"`
}

func toProjectResponse(p *projects.Project) projectResponse {
	return projectResponse{
		ID:          p.ID(),
		Name:        p.Name(),
		Description: p.Description(),
		TaskCount:   p.TaskCount(),
		CreatedAt:   p.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   p.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ── Handler ───────────────────────────────────────────────────────────────────

type ProjectHandler struct {
	projects projectService
	tasks    taskLister
}

func NewProjectHandler(p projectService, t taskLister) *ProjectHandler {
	return &ProjectHandler{projects: p, tasks: t}
}

func (h *ProjectHandler) Create(c echo.Context) error {
	var req createProjectRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	project, err := h.projects.Create(c.Request().Context(), uuid.NewString(), req.Name, req.Description)
	if err != nil {
		return h.projectError(err)
	}
	return c.JSON(http.StatusCreated, toProjectResponse(project))
}

func (h *ProjectHandler) Rename(c echo.Context) error {
	var req renameProjectRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	project, err := h.projects.Rename(c.Request().Context(), c.Param("id"), req.Name)
	if err != nil {
		return h.projectError(err)
	}
	return c.JSON(http.StatusOK, toProjectResponse(project))
}

func (h *ProjectHandler) Delete(c echo.Context) error {
	if err := h.projects.Delete(c.Request().Context(), c.Param("id")); err != nil {
		return h.projectError(err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *ProjectHandler) GetByID(c echo.Context) error {
	project, err := h.projects.GetByID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return h.projectError(err)
	}
	return c.JSON(http.StatusOK, toProjectResponse(project))
}

// GetWithTasks is the canonical Rule 5 example: the handler calls both services
// independently. No aggregate imports the other.
func (h *ProjectHandler) GetWithTasks(c echo.Context) error {
	id := c.Param("id")

	project, err := h.projects.GetByID(c.Request().Context(), id)
	if err != nil {
		return h.projectError(err)
	}

	taskList, err := h.tasks.ListByProject(c.Request().Context(), id)
	if err != nil {
		slog.Error("failed to list tasks for project", "project_id", id, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	taskResponses := make([]taskResponse, len(taskList))
	for i, t := range taskList {
		taskResponses[i] = toTaskResponse(t)
	}

	return c.JSON(http.StatusOK, projectWithTasksResponse{
		projectResponse: toProjectResponse(project),
		Tasks:           taskResponses,
	})
}

func (h *ProjectHandler) ListAll(c echo.Context) error {
	all, err := h.projects.ListAll(c.Request().Context())
	if err != nil {
		slog.Error("failed to list projects", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
	resp := make([]projectResponse, len(all))
	for i, p := range all {
		resp[i] = toProjectResponse(p)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *ProjectHandler) projectError(err error) error {
	switch {
	case errors.Is(err, projects.ErrNotFound):
		slog.Warn("project not found", "error", err)
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, projects.ErrEmptyName), errors.Is(err, projects.ErrNameTooLong):
		slog.Warn("project business rule violation", "error", err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	default:
		slog.Error("unexpected project error", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
}

// Compile-time guards.
var _ projectService = (*projects.Service)(nil)
var _ taskLister = (*tasks.Service)(nil)
