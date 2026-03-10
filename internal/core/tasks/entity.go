package tasks

import "time"

type Status string

const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

// Task is the core aggregate. No framework tags, no external imports.
type Task struct {
	id          string
	projectID   string // Primitive reference to the projects aggregate.
	title       string
	description string
	status      Status
	priority    Priority
	dueDate     *time.Time
	createdAt   time.Time
	updatedAt   time.Time
}

func NewTask(id, title, description, projectID string, priority Priority, dueDate *time.Time, now time.Time) (*Task, error) {
	if title == "" {
		return nil, ErrEmptyTitle
	}
	if len(title) > 200 {
		return nil, ErrTitleTooLong
	}
	return &Task{
		id:          id,
		projectID:   projectID,
		title:       title,
		description: description,
		status:      StatusPending,
		priority:    priority,
		dueDate:     dueDate,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

func (t *Task) Start(now time.Time) error {
	if t.status != StatusPending {
		return ErrInvalidTransition
	}
	t.status = StatusInProgress
	t.updatedAt = now
	return nil
}

func (t *Task) Complete(now time.Time) error {
	if t.status == StatusDone {
		return ErrAlreadyCompleted
	}
	t.status = StatusDone
	t.updatedAt = now
	return nil
}

func (t *Task) Reopen(now time.Time) error {
	if t.status != StatusDone {
		return ErrInvalidTransition
	}
	t.status = StatusPending
	t.updatedAt = now
	return nil
}

func (t *Task) UpdateTitle(title string, now time.Time) error {
	if title == "" {
		return ErrEmptyTitle
	}
	if len(title) > 200 {
		return ErrTitleTooLong
	}
	t.title = title
	t.updatedAt = now
	return nil
}

func (t *Task) ChangePriority(p Priority, now time.Time) {
	t.priority = p
	t.updatedAt = now
}

func (t *Task) IsOverdue(now time.Time) bool {
	if t.dueDate == nil || t.status == StatusDone {
		return false
	}
	return now.After(*t.dueDate)
}

// Getters — the only way for outer layers to read state.
func (t *Task) ID() string          { return t.id }
func (t *Task) ProjectID() string   { return t.projectID }
func (t *Task) Title() string       { return t.title }
func (t *Task) Description() string { return t.description }
func (t *Task) Status() Status      { return t.status }
func (t *Task) Priority() Priority  { return t.priority }
func (t *Task) DueDate() *time.Time { return t.dueDate }
func (t *Task) CreatedAt() time.Time { return t.createdAt }
func (t *Task) UpdatedAt() time.Time { return t.updatedAt }
