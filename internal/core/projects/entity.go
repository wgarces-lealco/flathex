package projects

import "time"

// Project is an aggregate root. It holds primitive task IDs only —
// it never imports the tasks package (Rule 5: Inter-Aggregate Isolation).
type Project struct {
	id          string
	name        string
	description string
	taskIDs     []string
	createdAt   time.Time
	updatedAt   time.Time
}

func NewProject(id, name, description string, now time.Time) (*Project, error) {
	if name == "" {
		return nil, ErrEmptyName
	}
	if len(name) > 100 {
		return nil, ErrNameTooLong
	}
	return &Project{
		id:          id,
		name:        name,
		description: description,
		taskIDs:     []string{},
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// RestoreProject reconstructs a Project loaded from persistence. Outbound adapters only.
func RestoreProject(id, name, description string, taskIDs []string, createdAt, updatedAt time.Time) *Project {
	ids := append([]string(nil), taskIDs...)
	return &Project{
		id: id, name: name, description: description, taskIDs: ids,
		createdAt: createdAt, updatedAt: updatedAt,
	}
}

func (p *Project) AddTask(taskID string, now time.Time) {
	if p.HasTask(taskID) {
		return
	}
	p.taskIDs = append(p.taskIDs, taskID)
	p.updatedAt = now
}

func (p *Project) RemoveTask(taskID string, now time.Time) {
	filtered := make([]string, 0, len(p.taskIDs))
	for _, id := range p.taskIDs {
		if id != taskID {
			filtered = append(filtered, id)
		}
	}
	p.taskIDs = filtered
	p.updatedAt = now
}

func (p *Project) HasTask(taskID string) bool {
	for _, id := range p.taskIDs {
		if id == taskID {
			return true
		}
	}
	return false
}

func (p *Project) TaskCount() int { return len(p.taskIDs) }

// Rename updates the project name.
func (p *Project) Rename(name string, now time.Time) error {
	if name == "" {
		return ErrEmptyName
	}
	if len(name) > 100 {
		return ErrNameTooLong
	}
	p.name = name
	p.updatedAt = now
	return nil
}

func (p *Project) ID() string          { return p.id }
func (p *Project) Name() string        { return p.name }
func (p *Project) Description() string { return p.description }
func (p *Project) TaskIDs() []string   { return append([]string{}, p.taskIDs...) }
func (p *Project) CreatedAt() time.Time { return p.createdAt }
func (p *Project) UpdatedAt() time.Time { return p.updatedAt }
