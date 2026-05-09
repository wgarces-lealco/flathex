package memory

import (
	"context"
	"flathex/internal/core/notifications"
	"log/slog"
)

// NoOpSender implements notifications.Sender for local development.
// It logs the notification instead of sending a real email.
type NoOpSender struct{}

func (NoOpSender) Send(_ context.Context, n *notifications.Notification) error {
	slog.Info("📬 [no-op sender] notification dispatched",
		"type", n.Type(),
		"recipient", n.Recipient(),
		"subject", n.Subject(),
	)
	return nil
}
