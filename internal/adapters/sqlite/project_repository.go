package sqlite

import (
	"context"
	"database/sql"
	"flathex/internal/core/projects"
	"time"
)

type ProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Save(ctx context.Context, p *projects.Project) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO projects (id, name, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			description = excluded.description,
			created_at = excluded.created_at,
			updated_at = excluded.updated_at`,
		p.ID(), p.Name(), p.Description(),
		p.CreatedAt().Format(time.RFC3339Nano),
		p.UpdatedAt().Format(time.RFC3339Nano),
	)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM project_tasks WHERE project_id = ?`, p.ID()); err != nil {
		return err
	}
	for _, tid := range p.TaskIDs() {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO project_tasks (project_id, task_id) VALUES (?, ?)`,
			p.ID(), tid); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *ProjectRepository) FindByID(ctx context.Context, id string) (*projects.Project, error) {
	var pid, name, description, createdAtStr, updatedAtStr string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, description, created_at, updated_at FROM projects WHERE id = ?`, id,
	).Scan(&pid, &name, &description, &createdAtStr, &updatedAtStr)
	if err == sql.ErrNoRows {
		return nil, projects.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	createdAt, err := time.Parse(time.RFC3339Nano, createdAtStr)
	if err != nil {
		return nil, err
	}
	updatedAt, err := time.Parse(time.RFC3339Nano, updatedAtStr)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT task_id FROM project_tasks WHERE project_id = ? ORDER BY task_id`, pid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var taskIDs []string
	for rows.Next() {
		var tid string
		if err := rows.Scan(&tid); err != nil {
			return nil, err
		}
		taskIDs = append(taskIDs, tid)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return projects.RestoreProject(pid, name, description, taskIDs, createdAt, updatedAt), nil
}

func (r *ProjectRepository) FindAll(ctx context.Context) ([]*projects.Project, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id FROM projects ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]*projects.Project, 0, len(ids))
	for _, id := range ids {
		p, err := r.FindByID(ctx, id)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func (r *ProjectRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return projects.ErrNotFound
	}
	return nil
}
