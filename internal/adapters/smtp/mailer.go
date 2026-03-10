//go:build ignore
// This file is a REFERENCE showing how an external I/O port is implemented.
// In production, pass this to notifications.NewService() instead of the no-op sender.

package smtp

import (
	"context"
	"flathex/internal/core/notifications"
	"fmt"
	"net/smtp"
)

type Mailer struct {
	host string
	port string
	from string
}

func NewMailer(host, port, from string) *Mailer {
	return &Mailer{host: host, port: port, from: from}
}

// Send implements notifications.Sender.
// The Anti-Corruption Layer is trivial here: we just read from the domain entity.
func (m *Mailer) Send(_ context.Context, n *notifications.Notification) error {
	addr := m.host + ":" + m.port
	msg := []byte(fmt.Sprintf(
		"To: %s\r\nSubject: %s\r\n\r\n%s",
		n.Recipient(), n.Subject(), n.Body(),
	))
	return smtp.SendMail(addr, nil, m.from, []string{n.Recipient()}, msg)
}
