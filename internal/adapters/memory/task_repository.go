package memory

import (
	"context"
	"flathex/internal/core/tasks"
	"sync"
)

// TaskRepository is both the in-memory adapter for the demo and the Fake
// used in tests. A single type serves two purposes by design.
type TaskRepository struct {
	mu    sync.RWMutex
	store map[string]*tasks.Task
}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{store: make(map[string]*tasks.Task)}
}

func (r *TaskRepository) Save(_ context.Context, t *tasks.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[t.ID()] = t
	return nil
}

func (r *TaskRepository) FindByID(_ context.Context, id string) (*tasks.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.store[id]
	if !ok {
		return nil, tasks.ErrNotFound
	}
	return t, nil
}

func (r *TaskRepository) FindByProject(_ context.Context, projectID string) ([]*tasks.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := []*tasks.Task{}
	for _, t := range r.store {
		if t.ProjectID() == projectID {
			result = append(result, t)
		}
	}
	return result, nil
}

func (r *TaskRepository) FindAll(_ context.Context) ([]*tasks.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*tasks.Task, 0, len(r.store))
	for _, t := range r.store {
		result = append(result, t)
	}
	return result, nil
}

func (r *TaskRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.store[id]; !ok {
		return tasks.ErrNotFound
	}
	delete(r.store, id)
	return nil
}
