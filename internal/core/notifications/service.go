package notifications

import (
	"context"
	"fmt"
)

type Service struct {
	sender Sender
	clock  Clock
}

func NewService(sender Sender, clock Clock) *Service {
	return &Service{sender: sender, clock: clock}
}

// NotifyTaskCompleted fires when a task transitions to Done.
// The handler passes primitive values — no Task struct crosses boundaries.
func (s *Service) NotifyTaskCompleted(ctx context.Context, id, recipient, taskTitle string) error {
	n, err := NewNotification(
		id,
		recipient,
		TypeTaskCompleted,
		"Task completed: "+taskTitle,
		fmt.Sprintf("Great news! The task %q has been marked as done.", taskTitle),
		s.clock.Now(),
	)
	if err != nil {
		return err
	}
	return s.sender.Send(ctx, n)
}

// NotifyTaskOverdue fires when a task's due date is past and it is not done.
func (s *Service) NotifyTaskOverdue(ctx context.Context, id, recipient, taskTitle string) error {
	n, err := NewNotification(
		id,
		recipient,
		TypeTaskOverdue,
		"Overdue task: "+taskTitle,
		fmt.Sprintf("The task %q is overdue. Please review it.", taskTitle),
		s.clock.Now(),
	)
	if err != nil {
		return err
	}
	return s.sender.Send(ctx, n)
}

// NotifyProjectDone fires when every task in a project is completed.
func (s *Service) NotifyProjectDone(ctx context.Context, id, recipient, projectName string) error {
	n, err := NewNotification(
		id,
		recipient,
		TypeProjectDone,
		"Project completed: "+projectName,
		fmt.Sprintf("All tasks in project %q are done. Congratulations!", projectName),
		s.clock.Now(),
	)
	if err != nil {
		return err
	}
	return s.sender.Send(ctx, n)
}
