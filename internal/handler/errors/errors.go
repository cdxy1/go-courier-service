package errors

import "errors"

var (
	ErrBadRequest = errors.New("invalid request body")
)
