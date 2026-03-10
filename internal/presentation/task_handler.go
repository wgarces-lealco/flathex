package presentation

import (
	"context"
	"errors"
	"flathex/internal/core/notifications"
	"flathex/internal/core/projects"
	"flathex/internal/core/tasks"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ── Consumer-defined interfaces (Rule 1) ─────────────────────────────────────

type taskService interface {
	Create(ctx context.Context, id, title, description, projectID string, priority tasks.Priority, dueDate *time.Time) (*tasks.Task, error)
	Start(ctx context.Context, id string) (*tasks.Task, error)
	Complete(ctx context.Context, id string) (*tasks.Task, error)
	Reopen(ctx context.Context, id string) (*tasks.Task, error)
	UpdateTitle(ctx context.Context, id, title string) (*tasks.Task, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*tasks.Task, error)
	ListAll(ctx context.Context) ([]*tasks.Task, error)
}

type projectLinker interface {
	AddTask(ctx context.Context, projectID, taskID string) error
	RemoveTask(ctx context.Context, projectID, taskID string) error
	GetByID(ctx context.Context, id string) (*projects.Project, error)
}

type notifier interface {
	NotifyTaskCompleted(ctx context.Context, id, recipient, taskTitle string) error
}

// ── DTOs ─────────────────────────────────────────────────────────────────────
// json tags live here, never in core. (Rule 3: Anti-Corruption Layer)

type createTaskRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	ProjectID   string  `json:"project_id"`
	Priority    string  `json:"priority"`
	DueDate     *string `json:"due_date"` // ISO 8601: "2026-12-31"
}

type taskResponse struct {
	ID          string  `json:"id"`
	ProjectID   string  `json:"project_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	Priority    string  `json:"priority"`
	DueDate     *string `json:"due_date,omitempty"`
	Overdue     bool    `json:"overdue"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func toTaskResponse(t *tasks.Task) taskResponse {
	resp := taskResponse{
		ID:          t.ID(),
		ProjectID:   t.ProjectID(),
		Title:       t.Title(),
		Description: t.Description(),
		Status:      string(t.Status()),
		Priority:    string(t.Priority()),
		Overdue:     t.IsOverdue(time.Now()),
		CreatedAt:   t.CreatedAt().Format(time.RFC3339),
		UpdatedAt:   t.UpdatedAt().Format(time.RFC3339),
	}
	if t.DueDate() != nil {
		s := t.DueDate().Format("2006-01-02")
		resp.DueDate = &s
	}
	return resp
}

// ── Handler ───────────────────────────────────────────────────────────────────

type TaskHandler struct {
	tasks    taskService
	projects projectLinker
	notifs   notifier
}

func NewTaskHandler(t taskService, p projectLinker, n notifier) *TaskHandler {
	return &TaskHandler{tasks: t, projects: p, notifs: n}
}

func (h *TaskHandler) Create(c echo.Context) error {
	var req createTaskRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	priority := tasks.Priority(req.Priority)
	if priority == "" {
		priority = tasks.PriorityMedium
	}

	var dueDate *time.Time
	if req.DueDate != nil {
		t, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid due_date format, use YYYY-MM-DD")
		}
		dueDate = &t
	}

	task, err := h.tasks.Create(c.Request().Context(), uuid.NewString(), req.Title, req.Description, req.ProjectID, priority, dueDate)
	if err != nil {
		return h.taskError(err)
	}

	// Rule 5: inter-aggregate orchestration happens here, in the handler.
	// The tasks service and projects service never import each other.
	if req.ProjectID != "" {
		if _, projErr := h.projects.GetByID(c.Request().Context(), req.ProjectID); projErr == nil {
			if linkErr := h.projects.AddTask(c.Request().Context(), req.ProjectID, task.ID()); linkErr != nil {
				slog.Warn("could not link task to project", "task_id", task.ID(), "project_id", req.ProjectID, "error", linkErr)
			}
		}
	}

	return c.JSON(http.StatusCreated, toTaskResponse(task))
}

func (h *TaskHandler) Start(c echo.Context) error {
	task, err := h.tasks.Start(c.Request().Context(), c.Param("id"))
	if err != nil {
		return h.taskError(err)
	}
	return c.JSON(http.StatusOK, toTaskResponse(task))
}

func (h *TaskHandler) Complete(c echo.Context) error {
	task, err := h.tasks.Complete(c.Request().Context(), c.Param("id"))
	if err != nil {
		return h.taskError(err)
	}

	// Rule 5: cross-aggregate notification orchestrated here, not inside core.
	if email := c.QueryParam("notify"); email != "" {
		if err := h.notifs.NotifyTaskCompleted(c.Request().Context(), uuid.NewString(), email, task.Title()); err != nil {
			slog.Warn("notification failed after task completion", "task_id", c.Param("id"), "error", err)
		}
	}

	return c.JSON(http.StatusOK, toTaskResponse(task))
}

func (h *TaskHandler) Reopen(c echo.Context) error {
	task, err := h.tasks.Reopen(c.Request().Context(), c.Param("id"))
	if err != nil {
		return h.taskError(err)
	}
	return c.JSON(http.StatusOK, toTaskResponse(task))
}

func (h *TaskHandler) Delete(c echo.Context) error {
	if err := h.tasks.Delete(c.Request().Context(), c.Param("id")); err != nil {
		return h.taskError(err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *TaskHandler) GetByID(c echo.Context) error {
	task, err := h.tasks.GetByID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return h.taskError(err)
	}
	return c.JSON(http.StatusOK, toTaskResponse(task))
}

func (h *TaskHandler) ListAll(c echo.Context) error {
	all, err := h.tasks.ListAll(c.Request().Context())
	if err != nil {
		slog.Error("failed to list tasks", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
	resp := make([]taskResponse, len(all))
	for i, t := range all {
		resp[i] = toTaskResponse(t)
	}
	return c.JSON(http.StatusOK, resp)
}

// taskError translates domain sentinel errors → HTTP errors.
// The core is HTTP-agnostic; this layer owns the translation (Rule 3).
func (h *TaskHandler) taskError(err error) error {
	switch {
	case errors.Is(err, tasks.ErrNotFound):
		slog.Warn("task not found", "error", err)
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, tasks.ErrEmptyTitle),
		errors.Is(err, tasks.ErrTitleTooLong),
		errors.Is(err, tasks.ErrInvalidTransition),
		errors.Is(err, tasks.ErrAlreadyCompleted):
		slog.Warn("task business rule violation", "error", err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	default:
		slog.Error("unexpected task error", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
}

// Compile-time guard: verifies notifications.Service satisfies our local notifier.
var _ notifier = (*notifications.Service)(nil)
