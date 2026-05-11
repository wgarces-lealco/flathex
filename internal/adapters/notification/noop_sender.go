package notification

import (
	"context"
	"log/slog"

	"flathex/internal/core/notifications"
)

// NoOpSender implements notifications.Sender for local development.
// It logs the notification instead of dispatching a real message.
type NoOpSender struct{}

func (NoOpSender) Send(_ context.Context, n *notifications.Notification) error {
	slog.Info("📬 [no-op sender] notification dispatched",
		"type", n.Type(),
		"recipient", n.Recipient(),
		"subject", n.Subject(),
	)
	return nil
}
