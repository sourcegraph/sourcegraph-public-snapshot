package redispool

import (
	"context"
	"sync"
)

// MemoryKeyValue returns an in memory KeyValue.
func MemoryKeyValue() KeyValue {
	var mu sync.Mutex
	m := map[string]NaiveValue{}
	store := func(_ context.Context, key string, f NaiveUpdater) error {
		mu.Lock()
		defer mu.Unlock()
		before, found := m[key]
		after, remove := f(before, found)
		if remove {
			if found {
				delete(m, key)
			}
		} else if before != after {
			m[key] = after
		}
		return nil
	}

	return FromNaiveKeyValueStore(store)
}
