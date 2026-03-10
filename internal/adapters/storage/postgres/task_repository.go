//go:build ignore

// This file is a REFERENCE showing the Anti-Corruption Layer pattern (Rule 3).
// It compiles with: go build -tags postgres
// In production, swap adapters/memory.TaskRepository for this in main.go.
package postgres

import (
	"context"
	"errors"
	"flathex/internal/core/tasks"
	"time"

	"gorm.io/gorm"
)

// taskModel is the DB representation. Tags live here, NEVER in core.
type taskModel struct {
	ID          string `gorm:"primaryKey"`
	ProjectID   string `gorm:"index"`
	Title       string `gorm:"not null"`
	Description string
	Status      string `gorm:"not null;default:'pending'"`
	Priority    string `gorm:"not null;default:'medium'"`
	DueDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"` // soft-delete for free
}

func (taskModel) TableName() string { return "tasks" }

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Save(ctx context.Context, t *tasks.Task) error {
	model := toModel(t)
	return r.db.WithContext(ctx).Save(&model).Error
}

func (r *TaskRepository) FindByID(ctx context.Context, id string) (*tasks.Task, error) {
	var model taskModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, tasks.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDomain(&model), nil
}

func (r *TaskRepository) FindByProject(ctx context.Context, projectID string) ([]*tasks.Task, error) {
	var models []taskModel
	if err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&models).Error; err != nil {
		return nil, err
	}
	return toDomainSlice(models), nil
}

func (r *TaskRepository) FindAll(ctx context.Context) ([]*tasks.Task, error) {
	var models []taskModel
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}
	return toDomainSlice(models), nil
}

func (r *TaskRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&taskModel{}, "id = ?", id).Error
}

// toModel maps a pure domain entity → DB model (Anti-Corruption Layer).
func toModel(t *tasks.Task) taskModel {
	return taskModel{
		ID:          t.ID(),
		ProjectID:   t.ProjectID(),
		Title:       t.Title(),
		Description: t.Description(),
		Status:      string(t.Status()),
		Priority:    string(t.Priority()),
		DueDate:     t.DueDate(),
		CreatedAt:   t.CreatedAt(),
		UpdatedAt:   t.UpdatedAt(),
	}
}

// toDomain maps a DB model → pure domain entity (Anti-Corruption Layer).
func toDomain(m *taskModel) *tasks.Task {
	due := m.DueDate
	t, _ := tasks.NewTask(m.ID, m.Title, m.Description, m.ProjectID,
		tasks.Priority(m.Priority), due, m.CreatedAt)
	return t
}

func toDomainSlice(models []taskModel) []*tasks.Task {
	result := make([]*tasks.Task, len(models))
	for i := range models {
		result[i] = toDomain(&models[i])
	}
	return result
}
