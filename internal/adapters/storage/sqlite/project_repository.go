package sqlite

import (
	"context"
	"errors"
	"time"

	"flathex/internal/core/projects"

	"gorm.io/gorm"
)

// projectModel is the GORM persistence model for the projects aggregate.
type projectModel struct {
	ID          string `gorm:"primaryKey"`
	Name        string `gorm:"not null"`
	Description string `gorm:"not null;default:''"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (projectModel) TableName() string { return "projects" }

// projectTaskModel models the many-to-many join table between projects and tasks.
type projectTaskModel struct {
	ProjectID string `gorm:"primaryKey;not null"`
	TaskID    string `gorm:"primaryKey;not null"`
}

func (projectTaskModel) TableName() string { return "project_tasks" }

// ProjectRepository implements projects.Repository using GORM.
type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Save(ctx context.Context, p *projects.Project) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		m := projectModel{
			ID:          p.ID(),
			Name:        p.Name(),
			Description: p.Description(),
			CreatedAt:   p.CreatedAt(),
			UpdatedAt:   p.UpdatedAt(),
		}
		if err := tx.Save(&m).Error; err != nil {
			return err
		}

		if err := tx.Delete(&projectTaskModel{}, "project_id = ?", p.ID()).Error; err != nil {
			return err
		}
		for _, tid := range p.TaskIDs() {
			if err := tx.Create(&projectTaskModel{ProjectID: p.ID(), TaskID: tid}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *ProjectRepository) FindByID(ctx context.Context, id string) (*projects.Project, error) {
	var m projectModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, projects.ErrNotFound
		}
		return nil, err
	}

	var joins []projectTaskModel
	if err := r.db.WithContext(ctx).
		Where("project_id = ?", id).
		Order("task_id").
		Find(&joins).Error; err != nil {
		return nil, err
	}

	taskIDs := make([]string, 0, len(joins))
	for _, j := range joins {
		taskIDs = append(taskIDs, j.TaskID)
	}

	return toProject(m, taskIDs), nil
}

func (r *ProjectRepository) FindAll(ctx context.Context) ([]*projects.Project, error) {
	var rows []projectModel
	if err := r.db.WithContext(ctx).Order("created_at").Find(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]*projects.Project, 0, len(rows))
	for _, m := range rows {
		var joins []projectTaskModel
		if err := r.db.WithContext(ctx).
			Where("project_id = ?", m.ID).
			Order("task_id").
			Find(&joins).Error; err != nil {
			return nil, err
		}
		taskIDs := make([]string, 0, len(joins))
		for _, j := range joins {
			taskIDs = append(taskIDs, j.TaskID)
		}
		out = append(out, toProject(m, taskIDs))
	}
	return out, nil
}

func (r *ProjectRepository) Delete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Delete(&projectModel{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return projects.ErrNotFound
	}
	return nil
}

// --- mapping helpers (Anti-Corruption Layer) ---

func toProject(m projectModel, taskIDs []string) *projects.Project {
	return projects.RestoreProject(
		m.ID, m.Name, m.Description,
		taskIDs,
		m.CreatedAt, m.UpdatedAt,
	)
}

