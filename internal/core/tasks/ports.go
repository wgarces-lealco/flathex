package tasks

import (
	"context"
	"time"
)

// Repository is the outbound port for task persistence.
// Implemented by adapters (e.g. sqlite, memory for tests).
type Repository interface {
	Save(ctx context.Context, t *Task) error
	FindByID(ctx context.Context, id string) (*Task, error)
	FindByProject(ctx context.Context, projectID string) ([]*Task, error)
	FindAll(ctx context.Context) ([]*Task, error)
	Delete(ctx context.Context, id string) error
}

// Clock is an outbound port for time, making the service fully testable
// without depending on real wall-clock time.
type Clock interface {
	Now() time.Time
}
