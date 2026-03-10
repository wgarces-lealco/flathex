package tasks

import "errors"

var (
	ErrNotFound         = errors.New("task not found")
	ErrEmptyTitle       = errors.New("task title cannot be empty")
	ErrTitleTooLong     = errors.New("task title cannot exceed 200 characters")
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrAlreadyCompleted = errors.New("task is already completed")
)
