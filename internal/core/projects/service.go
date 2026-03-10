package projects

import "context"

type Service struct {
	repo  Repository
	clock Clock
}

func NewService(repo Repository, clock Clock) *Service {
	return &Service{repo: repo, clock: clock}
}

func (s *Service) Create(ctx context.Context, id, name, description string) (*Project, error) {
	project, err := NewProject(id, name, description, s.clock.Now())
	if err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, project); err != nil {
		return nil, err
	}
	return project, nil
}

// AddTask links a task (by primitive ID) to this project.
// The handler is responsible for ensuring the task actually exists — this
// service intentionally does NOT import core/tasks (Rule 5).
func (s *Service) AddTask(ctx context.Context, projectID, taskID string) error {
	project, err := s.repo.FindByID(ctx, projectID)
	if err != nil {
		return err
	}
	project.AddTask(taskID, s.clock.Now())
	return s.repo.Save(ctx, project)
}

func (s *Service) RemoveTask(ctx context.Context, projectID, taskID string) error {
	project, err := s.repo.FindByID(ctx, projectID)
	if err != nil {
		return err
	}
	project.RemoveTask(taskID, s.clock.Now())
	return s.repo.Save(ctx, project)
}

func (s *Service) Rename(ctx context.Context, id, name string) (*Project, error) {
	project, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := project.Rename(name, s.clock.Now()); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, project); err != nil {
		return nil, err
	}
	return project, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Project, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) ListAll(ctx context.Context) ([]*Project, error) {
	return s.repo.FindAll(ctx)
}
