package courier

import "errors"

var (
	ErrInvalidID    = errors.New("invalid id")
	ErrInvalidPhone = errors.New("invalid phone")
	ErrInvalidName  = errors.New("invalid name")
)
