package projects_test

import (
	"context"
	"flathex/internal/core/projects"
	"testing"
	"time"
)

// ── Fakes ─────────────────────────────────────────────────────────────────────

type fakeRepo struct {
	store map[string]*projects.Project
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{store: make(map[string]*projects.Project)}
}

func (r *fakeRepo) Save(_ context.Context, p *projects.Project) error {
	r.store[p.ID()] = p
	return nil
}

func (r *fakeRepo) FindByID(_ context.Context, id string) (*projects.Project, error) {
	p, ok := r.store[id]
	if !ok {
		return nil, projects.ErrNotFound
	}
	return p, nil
}

func (r *fakeRepo) FindAll(_ context.Context) ([]*projects.Project, error) {
	result := make([]*projects.Project, 0, len(r.store))
	for _, p := range r.store {
		result = append(result, p)
	}
	return result, nil
}

func (r *fakeRepo) Delete(_ context.Context, id string) error {
	if _, ok := r.store[id]; !ok {
		return projects.ErrNotFound
	}
	delete(r.store, id)
	return nil
}

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

var clock = fixedClock{t: time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)}

func newSvc() *projects.Service {
	return projects.NewService(newFakeRepo(), clock)
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestCreate(t *testing.T) {
	tt := []struct {
		name    string
		pName   string
		wantErr error
	}{
		{name: "valid project", pName: "Flathex Demo", wantErr: nil},
		{name: "empty name", pName: "", wantErr: projects.ErrEmptyName},
		{name: "name too long", pName: string(make([]byte, 101)), wantErr: projects.ErrNameTooLong},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			svc := newSvc()
			p, err := svc.Create(context.Background(), "p-1", tc.pName, "")
			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.Name() != tc.pName {
				t.Errorf("expected name %q, got %q", tc.pName, p.Name())
			}
			if p.TaskCount() != 0 {
				t.Errorf("new project should have 0 tasks")
			}
		})
	}
}

func TestAddAndRemoveTask(t *testing.T) {
	tt := []struct {
		name         string
		taskIDs      []string
		removeIDs    []string
		wantCount    int
	}{
		{name: "add 3 tasks", taskIDs: []string{"t1", "t2", "t3"}, wantCount: 3},
		{name: "add duplicate is idempotent", taskIDs: []string{"t1", "t1", "t1"}, wantCount: 1},
		{name: "add then remove", taskIDs: []string{"t1", "t2"}, removeIDs: []string{"t1"}, wantCount: 1},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			svc := newSvc()
			svc.Create(context.Background(), "p-1", "Project", "")

			for _, id := range tc.taskIDs {
				if err := svc.AddTask(context.Background(), "p-1", id); err != nil {
					t.Fatalf("AddTask failed: %v", err)
				}
			}
			for _, id := range tc.removeIDs {
				if err := svc.RemoveTask(context.Background(), "p-1", id); err != nil {
					t.Fatalf("RemoveTask failed: %v", err)
				}
			}

			p, _ := svc.GetByID(context.Background(), "p-1")
			if p.TaskCount() != tc.wantCount {
				t.Errorf("expected %d tasks, got %d", tc.wantCount, p.TaskCount())
			}
		})
	}
}

func TestRename(t *testing.T) {
	tt := []struct {
		name    string
		newName string
		wantErr error
	}{
		{name: "valid rename", newName: "New Name", wantErr: nil},
		{name: "empty name", newName: "", wantErr: projects.ErrEmptyName},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			svc := newSvc()
			svc.Create(context.Background(), "p-1", "Original", "")

			_, err := svc.Rename(context.Background(), "p-1", tc.newName)
			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			p, _ := svc.GetByID(context.Background(), "p-1")
			if p.Name() != tc.newName {
				t.Errorf("expected %q, got %q", tc.newName, p.Name())
			}
		})
	}
}
