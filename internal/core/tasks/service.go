package tasks

import (
	"context"
	"time"
)

type Service struct {
	repo  Repository
	clock Clock
}

func NewService(repo Repository, clock Clock) *Service {
	return &Service{repo: repo, clock: clock}
}

func (s *Service) Create(ctx context.Context, id, title, description, projectID string, priority Priority, dueDate *time.Time) (*Task, error) {
	task, err := NewTask(id, title, description, projectID, priority, dueDate, s.clock.Now())
	if err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *Service) Start(ctx context.Context, id string) (*Task, error) {
	task, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := task.Start(s.clock.Now()); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *Service) Complete(ctx context.Context, id string) (*Task, error) {
	task, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := task.Complete(s.clock.Now()); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *Service) Reopen(ctx context.Context, id string) (*Task, error) {
	task, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := task.Reopen(s.clock.Now()); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *Service) UpdateTitle(ctx context.Context, id, title string) (*Task, error) {
	task, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := task.UpdateTitle(title, s.clock.Now()); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Task, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) ListByProject(ctx context.Context, projectID string) ([]*Task, error) {
	return s.repo.FindByProject(ctx, projectID)
}

func (s *Service) ListAll(ctx context.Context) ([]*Task, error) {
	return s.repo.FindAll(ctx)
}
