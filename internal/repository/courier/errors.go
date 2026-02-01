package courier

import "errors"

var (
	ErrPhoneExists      = errors.New("courier with this phone already exists")
	ErrCourierNotFound  = errors.New("courier not found")
	ErrDatabaseInternal = errors.New("database error")
	ErrReadingData      = errors.New("error reading data")
)
