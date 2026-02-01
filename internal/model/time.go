package model

import "time"

type NowFunc func() time.Time

func UTCNow() time.Time {
	return time.Now().UTC()
}
