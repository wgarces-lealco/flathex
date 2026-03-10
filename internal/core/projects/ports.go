package projects

import (
	"context"
	"time"
)

type Repository interface {
	Save(ctx context.Context, p *Project) error
	FindByID(ctx context.Context, id string) (*Project, error)
	FindAll(ctx context.Context) ([]*Project, error)
	Delete(ctx context.Context, id string) error
}

type Clock interface {
	Now() time.Time
}
