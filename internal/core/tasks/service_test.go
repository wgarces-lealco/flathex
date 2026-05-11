package tasks_test

import (
	"context"
	"flathex/internal/core/tasks"
	"testing"
	"time"
)

// ── Fakes (in-memory) ─────────────────────────────────────────────────────────
// These live in the test file, not in adapters/storage/memory, to keep test ownership
// local. For integration tests, use adapters/storage/memory directly.

type fakeRepo struct {
	store map[string]*tasks.Task
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{store: make(map[string]*tasks.Task)}
}

func (r *fakeRepo) Save(_ context.Context, t *tasks.Task) error {
	r.store[t.ID()] = t
	return nil
}

func (r *fakeRepo) FindByID(_ context.Context, id string) (*tasks.Task, error) {
	t, ok := r.store[id]
	if !ok {
		return nil, tasks.ErrNotFound
	}
	return t, nil
}

func (r *fakeRepo) FindByProject(_ context.Context, projectID string) ([]*tasks.Task, error) {
	var result []*tasks.Task
	for _, t := range r.store {
		if t.ProjectID() == projectID {
			result = append(result, t)
		}
	}
	return result, nil
}

func (r *fakeRepo) FindAll(_ context.Context) ([]*tasks.Task, error) {
	result := make([]*tasks.Task, 0, len(r.store))
	for _, t := range r.store {
		result = append(result, t)
	}
	return result, nil
}

func (r *fakeRepo) Delete(_ context.Context, id string) error {
	if _, ok := r.store[id]; !ok {
		return tasks.ErrNotFound
	}
	delete(r.store, id)
	return nil
}

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

var (
	fixedNow = time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	clock    = fixedClock{t: fixedNow}
)

func newSvc() *tasks.Service {
	return tasks.NewService(newFakeRepo(), clock)
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestCreate(t *testing.T) {
	tt := []struct {
		name        string
		title       string
		priority    tasks.Priority
		wantErr     error
		wantNilTask bool
	}{
		{name: "valid task — high priority", title: "Write tests", priority: tasks.PriorityHigh, wantErr: nil, wantNilTask: false},
		{name: "valid task — medium priority", title: "Write tests", priority: tasks.PriorityMedium, wantErr: nil, wantNilTask: false},
		{name: "valid task — low priority", title: "Write tests", priority: tasks.PriorityLow, wantErr: nil, wantNilTask: false},
		{name: "empty title", title: "", priority: tasks.PriorityMedium, wantErr: tasks.ErrEmptyTitle, wantNilTask: true},
		{name: "title too long", title: string(make([]byte, 201)), priority: tasks.PriorityMedium, wantErr: tasks.ErrTitleTooLong, wantNilTask: true},
		{name: "invalid priority", title: "Write tests", priority: tasks.Priority("urgent"), wantErr: tasks.ErrInvalidPriority, wantNilTask: true},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			svc := newSvc()
			task, err := svc.Create(context.Background(), "id-1", tc.title, "", "", tc.priority, nil)
			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantNilTask && task != nil {
				t.Fatal("expected nil task")
			}
			if task.Status() != tasks.StatusPending {
				t.Errorf("new task should be pending, got %s", task.Status())
			}
		})
	}
}

func TestStatusTransitions(t *testing.T) {
	tt := []struct {
		name      string
		actions   []func(*tasks.Service, string) error
		wantFinal tasks.Status
		wantErr   error
	}{
		{
			name: "pending → in_progress → done",
			actions: []func(*tasks.Service, string) error{
				func(s *tasks.Service, id string) error { _, err := s.Start(context.Background(), id); return err },
				func(s *tasks.Service, id string) error { _, err := s.Complete(context.Background(), id); return err },
			},
			wantFinal: tasks.StatusDone,
		},
		{
			name: "cannot complete without starting first",
			actions: []func(*tasks.Service, string) error{
				func(s *tasks.Service, id string) error { _, err := s.Complete(context.Background(), id); return err },
			},
			wantErr: tasks.ErrInvalidTransition,
		},
		{
			name: "done → pending (reopen)",
			actions: []func(*tasks.Service, string) error{
				func(s *tasks.Service, id string) error { _, err := s.Start(context.Background(), id); return err },
				func(s *tasks.Service, id string) error { _, err := s.Complete(context.Background(), id); return err },
				func(s *tasks.Service, id string) error { _, err := s.Reopen(context.Background(), id); return err },
			},
			wantFinal: tasks.StatusPending,
		},
		{
			name: "cannot start a done task",
			actions: []func(*tasks.Service, string) error{
				func(s *tasks.Service, id string) error { _, err := s.Start(context.Background(), id); return err },
				func(s *tasks.Service, id string) error { _, err := s.Complete(context.Background(), id); return err },
				func(s *tasks.Service, id string) error { _, err := s.Start(context.Background(), id); return err },
			},
			wantErr: tasks.ErrInvalidTransition,
		},
		{
			name: "cannot complete twice",
			actions: []func(*tasks.Service, string) error{
				func(s *tasks.Service, id string) error { _, err := s.Start(context.Background(), id); return err },
				func(s *tasks.Service, id string) error { _, err := s.Complete(context.Background(), id); return err },
				func(s *tasks.Service, id string) error { _, err := s.Complete(context.Background(), id); return err },
			},
			wantErr: tasks.ErrInvalidTransition,
		},
		{
			name: "pending → cancelled",
			actions: []func(*tasks.Service, string) error{
				func(s *tasks.Service, id string) error { _, err := s.Cancel(context.Background(), id); return err },
			},
			wantFinal: tasks.StatusCancelled,
		},
		{
			name: "in_progress → cancelled",
			actions: []func(*tasks.Service, string) error{
				func(s *tasks.Service, id string) error { _, err := s.Start(context.Background(), id); return err },
				func(s *tasks.Service, id string) error { _, err := s.Cancel(context.Background(), id); return err },
			},
			wantFinal: tasks.StatusCancelled,
		},
		{
			name: "cancelled → pending (reopen)",
			actions: []func(*tasks.Service, string) error{
				func(s *tasks.Service, id string) error { _, err := s.Cancel(context.Background(), id); return err },
				func(s *tasks.Service, id string) error { _, err := s.Reopen(context.Background(), id); return err },
			},
			wantFinal: tasks.StatusPending,
		},
		{
			name: "cannot cancel a done task",
			actions: []func(*tasks.Service, string) error{
				func(s *tasks.Service, id string) error { _, err := s.Start(context.Background(), id); return err },
				func(s *tasks.Service, id string) error { _, err := s.Complete(context.Background(), id); return err },
				func(s *tasks.Service, id string) error { _, err := s.Cancel(context.Background(), id); return err },
			},
			wantErr: tasks.ErrInvalidTransition,
		},
		{
			name: "cannot cancel twice",
			actions: []func(*tasks.Service, string) error{
				func(s *tasks.Service, id string) error { _, err := s.Cancel(context.Background(), id); return err },
				func(s *tasks.Service, id string) error { _, err := s.Cancel(context.Background(), id); return err },
			},
			wantErr: tasks.ErrAlreadyCancelled,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			svc := newSvc()
			task, _ := svc.Create(context.Background(), "id-1", "My task", "", "", tasks.PriorityLow, nil)

			var lastErr error
			for _, action := range tc.actions {
				lastErr = action(svc, task.ID())
			}

			if tc.wantErr != nil {
				if lastErr == nil {
					t.Fatalf("expected error %v, got nil", tc.wantErr)
				}
				return
			}
			if lastErr != nil {
				t.Fatalf("unexpected error: %v", lastErr)
			}

			got, _ := svc.GetByID(context.Background(), task.ID())
			if got.Status() != tc.wantFinal {
				t.Errorf("expected status %s, got %s", tc.wantFinal, got.Status())
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tt := []struct {
		name    string
		id      string
		wantErr error
	}{
		{name: "existing task", id: "id-1", wantErr: nil},
		{name: "non-existing task", id: "ghost", wantErr: tasks.ErrNotFound},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			svc := newSvc()
			svc.Create(context.Background(), "id-1", "A task", "", "", tasks.PriorityHigh, nil)

			err := svc.Delete(context.Background(), tc.id)
			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tc.wantErr)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestIsOverdue(t *testing.T) {
	past := fixedNow.Add(-24 * time.Hour)
	future := fixedNow.Add(24 * time.Hour)

	tt := []struct {
		name        string
		dueDate     *time.Time
		complete    bool
		wantOverdue bool
	}{
		{name: "past due date, not done", dueDate: &past, complete: false, wantOverdue: true},
		{name: "future due date", dueDate: &future, complete: false, wantOverdue: false},
		{name: "past due date but done", dueDate: &past, complete: true, wantOverdue: false},
		{name: "no due date", dueDate: nil, complete: false, wantOverdue: false},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			svc := newSvc()
			task, _ := svc.Create(context.Background(), "id-1", "Task", "", "", tasks.PriorityLow, tc.dueDate)
		if tc.complete {
			svc.Start(context.Background(), task.ID())
			svc.Complete(context.Background(), task.ID())
		}
			got, _ := svc.GetByID(context.Background(), task.ID())
			if got.IsOverdue(fixedNow) != tc.wantOverdue {
				t.Errorf("IsOverdue: expected %v, got %v", tc.wantOverdue, got.IsOverdue(fixedNow))
			}
		})
	}
}
