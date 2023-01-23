package redispool

import (
	"context"
	"sync"
)

// MemoryKeyValue returns an in memory KeyValue.
func MemoryKeyValue() KeyValue {
	var mu sync.Mutex
	m := map[string]string{}
	store := func(_ context.Context, key string, f NaiveUpdater) error {
		mu.Lock()
		defer mu.Unlock()
		currentValue, found := m[key]
		newValue, remove := f(currentValue, found)
		if remove {
			if found {
				delete(m, key)
			}
		} else if currentValue != newValue {
			m[key] = newValue
		}
		return nil
	}

	return FromNaiveKeyValueStore(store)
}
