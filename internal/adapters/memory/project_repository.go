package memory

import (
	"context"
	"flathex/internal/core/projects"
	"sync"
)

type ProjectRepository struct {
	mu    sync.RWMutex
	store map[string]*projects.Project
}

func NewProjectRepository() *ProjectRepository {
	return &ProjectRepository{store: make(map[string]*projects.Project)}
}

func (r *ProjectRepository) Save(_ context.Context, p *projects.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[p.ID()] = p
	return nil
}

func (r *ProjectRepository) FindByID(_ context.Context, id string) (*projects.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.store[id]
	if !ok {
		return nil, projects.ErrNotFound
	}
	return p, nil
}

func (r *ProjectRepository) FindAll(_ context.Context) ([]*projects.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*projects.Project, 0, len(r.store))
	for _, p := range r.store {
		result = append(result, p)
	}
	return result, nil
}

func (r *ProjectRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.store[id]; !ok {
		return projects.ErrNotFound
	}
	delete(r.store, id)
	return nil
}
