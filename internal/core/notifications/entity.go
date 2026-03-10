package notifications

import "time"

type NotificationType string

const (
	TypeTaskCompleted NotificationType = "task_completed"
	TypeTaskOverdue   NotificationType = "task_overdue"
	TypeProjectDone   NotificationType = "project_done"
)

type Notification struct {
	id        string
	recipient string
	notifType NotificationType
	subject   string
	body      string
	sentAt    time.Time
}

func NewNotification(id, recipient string, notifType NotificationType, subject, body string, now time.Time) (*Notification, error) {
	if recipient == "" {
		return nil, ErrEmptyRecipient
	}
	return &Notification{
		id:        id,
		recipient: recipient,
		notifType: notifType,
		subject:   subject,
		body:      body,
		sentAt:    now,
	}, nil
}

func (n *Notification) ID() string               { return n.id }
func (n *Notification) Recipient() string        { return n.recipient }
func (n *Notification) Type() NotificationType   { return n.notifType }
func (n *Notification) Subject() string          { return n.subject }
func (n *Notification) Body() string             { return n.body }
func (n *Notification) SentAt() time.Time        { return n.sentAt }
