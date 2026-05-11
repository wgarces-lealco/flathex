package notifications

import "time"

type NotificationType string

const (
	TypeTaskCompleted NotificationType = "task_completed"
	TypeTaskOverdue   NotificationType = "task_overdue"
	TypeProjectDone   NotificationType = "project_done"
)

// Notification is a Value Object — it is defined entirely by its content,
// has no identity of its own, and is immutable after construction.
// It is never persisted: it is created and handed to a Sender in one operation.
type Notification struct {
	recipient string
	notifType NotificationType
	subject   string
	body      string
	sentAt    time.Time
}

func NewNotification(recipient string, notifType NotificationType, subject, body string, now time.Time) (*Notification, error) {
	if recipient == "" {
		return nil, ErrEmptyRecipient
	}
	return &Notification{
		recipient: recipient,
		notifType: notifType,
		subject:   subject,
		body:      body,
		sentAt:    now,
	}, nil
}

func (n *Notification) Recipient() string       { return n.recipient }
func (n *Notification) Type() NotificationType  { return n.notifType }
func (n *Notification) Subject() string         { return n.subject }
func (n *Notification) Body() string            { return n.body }
func (n *Notification) SentAt() time.Time       { return n.sentAt }
