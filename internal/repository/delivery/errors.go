package delivery

import "errors"

var (
	ErrDeliveryNotFound     = errors.New("delivery not found")
	ErrDeliveryTableMissing = errors.New("delivery table is missing")
	ErrDatabaseInternal     = errors.New("database error")
)
