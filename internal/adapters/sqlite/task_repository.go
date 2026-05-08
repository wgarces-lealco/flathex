package sqlite

import (
	"context"
	"database/sql"
	"flathex/internal/core/tasks"
	"time"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Save(ctx context.Context, t *tasks.Task) error {
	var dueVal any
	if d := t.DueDate(); d != nil {
		dueVal = d.Format("2006-01-02")
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO tasks (id, project_id, title, description, status, priority, due_date, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			project_id = excluded.project_id,
			title = excluded.title,
			description = excluded.description,
			status = excluded.status,
			priority = excluded.priority,
			due_date = excluded.due_date,
			created_at = excluded.created_at,
			updated_at = excluded.updated_at`,
		t.ID(), t.ProjectID(), t.Title(), t.Description(),
		string(t.Status()), string(t.Priority()),
		dueVal,
		t.CreatedAt().Format(time.RFC3339Nano),
		t.UpdatedAt().Format(time.RFC3339Nano),
	)
	return err
}

func (r *TaskRepository) FindByID(ctx context.Context, id string) (*tasks.Task, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, project_id, title, description, status, priority, due_date, created_at, updated_at
		FROM tasks WHERE id = ?`, id)
	return scanTask(row)
}

func (r *TaskRepository) FindByProject(ctx context.Context, projectID string) ([]*tasks.Task, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, project_id, title, description, status, priority, due_date, created_at, updated_at
		FROM tasks WHERE project_id = ? ORDER BY created_at`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*tasks.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *TaskRepository) FindAll(ctx context.Context) ([]*tasks.Task, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, project_id, title, description, status, priority, due_date, created_at, updated_at
		FROM tasks ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*tasks.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *TaskRepository) Delete(ctx context.Context, id string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM project_tasks WHERE task_id = ?`, id); err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM tasks WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return tasks.ErrNotFound
	}
	return tx.Commit()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTask(row rowScanner) (*tasks.Task, error) {
	var (
		id, projectID, title, description, status, priority string
		due                                                 sql.NullString
		createdAtStr, updatedAtStr                          string
	)
	if err := row.Scan(&id, &projectID, &title, &description, &status, &priority, &due, &createdAtStr, &updatedAtStr); err != nil {
		if err == sql.ErrNoRows {
			return nil, tasks.ErrNotFound
		}
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
	var duePtr *time.Time
	if due.Valid && due.String != "" {
		d, err := time.Parse("2006-01-02", due.String)
		if err != nil {
			return nil, err
		}
		duePtr = &d
	}
	return tasks.RestoreTask(id, projectID, title, description, tasks.Status(status), tasks.Priority(priority), duePtr, createdAt, updatedAt), nil
}
