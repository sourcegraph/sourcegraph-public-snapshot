package types

import "time"

const CurrentCacheVersion = 1

type BatchSpecExecutionCacheEntry struct {
	ID int64

	Key   string
	Value string

	Version int

	LastUsedAt time.Time
	CreatedAt  time.Time
}
