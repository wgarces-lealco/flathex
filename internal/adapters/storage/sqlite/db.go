package sqlite

import (
	"fmt"
	"os"
	"path/filepath"

	glebarez "github.com/glebarez/sqlite" // GORM driver — pure Go, no CGO (modernc.org/sqlite)
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open opens (or creates) a SQLite database at path using GORM.
// The schema is managed via AutoMigrate on the internal models.
func Open(path string) (*gorm.DB, error) {
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create db directory: %w", err)
		}
	}

	db, err := gorm.Open(glebarez.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("open sqlite %q: %w", path, err)
	}

	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	// AutoMigrate creates tables and adds missing columns. It never drops columns.
	if err := db.AutoMigrate(&taskModel{}, &projectModel{}, &projectTaskModel{}); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	return db, nil
}
