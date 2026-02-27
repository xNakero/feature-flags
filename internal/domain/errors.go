package domain

import "errors"

var (
	ErrNotFound      = errors.New("flag not found")
	ErrAlreadyExists = errors.New("flag already exists")
	ErrTypeMismatch  = errors.New("value type does not match flag type")
	ErrInvalidName   = errors.New("invalid flag name")
	ErrInvalidValue  = errors.New("invalid flag value")
)
