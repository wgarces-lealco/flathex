package sqlite

import (
	"context"
	"errors"
	"time"

	"flathex/internal/core/tasks"

	"gorm.io/gorm"
)

// taskModel is the GORM persistence model for the tasks aggregate.
// It lives only in this package; the core domain never sees it.
type taskModel struct {
	ID          string  `gorm:"primaryKey"`
	ProjectID   string  `gorm:"not null;default:''"`
	Title       string  `gorm:"not null"`
	Description string  `gorm:"not null;default:''"`
	Status      string  `gorm:"not null"`
	Priority    string  `gorm:"not null"`
	DueDate     *string // stored as "2006-01-02", nil when absent
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (taskModel) TableName() string { return "tasks" }

// TaskRepository implements tasks.Repository using GORM.
type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Save(ctx context.Context, t *tasks.Task) error {
	var dueVal *string
	if d := t.DueDate(); d != nil {
		s := d.Format("2006-01-02")
		dueVal = &s
	}
	m := taskModel{
		ID:          t.ID(),
		ProjectID:   t.ProjectID(),
		Title:       t.Title(),
		Description: t.Description(),
		Status:      string(t.Status()),
		Priority:    string(t.Priority()),
		DueDate:     dueVal,
		CreatedAt:   t.CreatedAt(),
		UpdatedAt:   t.UpdatedAt(),
	}
	return r.db.WithContext(ctx).Save(&m).Error
}

func (r *TaskRepository) FindByID(ctx context.Context, id string) (*tasks.Task, error) {
	var m taskModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, tasks.ErrNotFound
		}
		return nil, err
	}
	return toTask(m)
}

func (r *TaskRepository) FindByProject(ctx context.Context, projectID string) ([]*tasks.Task, error) {
	var rows []taskModel
	if err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("created_at").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return toTaskSlice(rows)
}

func (r *TaskRepository) FindAll(ctx context.Context) ([]*tasks.Task, error) {
	var rows []taskModel
	if err := r.db.WithContext(ctx).Order("created_at").Find(&rows).Error; err != nil {
		return nil, err
	}
	return toTaskSlice(rows)
}

func (r *TaskRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&projectTaskModel{}, "task_id = ?", id).Error; err != nil {
			return err
		}
		res := tx.Delete(&taskModel{}, "id = ?", id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return tasks.ErrNotFound
		}
		return nil
	})
}

// --- mapping helpers (Anti-Corruption Layer) ---

func toTask(m taskModel) (*tasks.Task, error) {
	var duePtr *time.Time
	if m.DueDate != nil && *m.DueDate != "" {
		d, err := time.Parse("2006-01-02", *m.DueDate)
		if err != nil {
			return nil, err
		}
		duePtr = &d
	}
	return tasks.RestoreTask(
		m.ID, m.ProjectID, m.Title, m.Description,
		tasks.Status(m.Status), tasks.Priority(m.Priority),
		duePtr, m.CreatedAt, m.UpdatedAt,
	), nil
}

func toTaskSlice(rows []taskModel) ([]*tasks.Task, error) {
	out := make([]*tasks.Task, 0, len(rows))
	for _, m := range rows {
		t, err := toTask(m)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, nil
}
