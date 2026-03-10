package notifications

import "errors"

var (
	ErrEmptyRecipient = errors.New("notification recipient cannot be empty")
	ErrSendFailed     = errors.New("failed to send notification")
)
