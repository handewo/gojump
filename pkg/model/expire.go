package model

import (
	"time"
)

type ExpireInfo struct {
	ExpireAt int64 `json:"expire_at"`
}

func (e *ExpireInfo) IsExpired(now time.Time) bool {
	if e.ExpireAt == 0 {
		return false
	}
	return e.ExpireAt < now.Unix()
}
