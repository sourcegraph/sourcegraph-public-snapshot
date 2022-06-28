package types

import (
	"encoding/json"
	"time"
)

const CurrentCacheVersion = 2

type BatchSpecExecutionCacheEntry struct {
	ID int64

	UserID int32

	Key   string
	Value string

	Version int

	LastUsedAt time.Time
	CreatedAt  time.Time
}

func NewCacheEntryFromResult(key string, result any) (*BatchSpecExecutionCacheEntry, error) {
	value, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	entry := &BatchSpecExecutionCacheEntry{Key: key, Value: string(value)}
	return entry, nil
}
