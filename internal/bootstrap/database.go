package bootstrap

import (
	"flathex/internal/adapters/sqlite"

	"gorm.io/gorm"
)

// OpenDB opens SQLite via GORM (pure Go, no CGO) and runs AutoMigrate.
func OpenDB(cfg Config) (*gorm.DB, error) {
	return sqlite.Open(cfg.SQLitePath)
}
