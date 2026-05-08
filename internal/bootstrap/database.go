package bootstrap

import (
	"database/sql"
	"flathex/internal/adapters/sqlite"
)

// OpenDB opens SQLite (schema migrated) from cfg.SQLitePath.
func OpenDB(cfg Config) (*sql.DB, error) {
	return sqlite.Open(cfg.SQLitePath)
}
