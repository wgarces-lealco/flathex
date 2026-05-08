package sqlite

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // driver name "sqlite"
)

var migrations = []string{
	`PRAGMA foreign_keys = ON`,
	`CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		project_id TEXT NOT NULL DEFAULT '',
		title TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL,
		priority TEXT NOT NULL,
		due_date TEXT,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS projects (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS project_tasks (
		project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
		task_id TEXT NOT NULL,
		PRIMARY KEY (project_id, task_id)
	)`,
	`CREATE INDEX IF NOT EXISTS idx_tasks_project ON tasks(project_id)`,
}

// Open opens SQLite at path, ensures parent directory exists, applies schema.
func Open(path string) (*sql.DB, error) {
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}

	db, err := sql.Open("sqlite", "file:"+path+"?_pragma=foreign_keys(ON)")
	if err != nil {
		return nil, err
	}
	for _, stmt := range migrations {
		if _, err := db.Exec(stmt); err != nil {
			_ = db.Close()
			return nil, err
		}
	}
	return db, nil
}
