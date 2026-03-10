package notifications

import (
	"context"
	"time"
)

// Sender is the outbound port for delivering notifications.
// Implemented by adapters/smtp (production) or a mock (tests).
type Sender interface {
	Send(ctx context.Context, n *Notification) error
}

type Clock interface {
	Now() time.Time
}
