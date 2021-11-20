package types

import "time"

type LifecycleHook struct {
	ID        int64
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiresAt time.Time
	URL       string
	Secret    string
}
