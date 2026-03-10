package projects

import "errors"

var (
	ErrNotFound   = errors.New("project not found")
	ErrEmptyName  = errors.New("project name cannot be empty")
	ErrNameTooLong = errors.New("project name cannot exceed 100 characters")
)
