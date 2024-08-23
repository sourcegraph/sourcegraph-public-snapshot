// Package memstore offers an in-memory store implementation for throttled.
package memstore // import "github.com/throttled/throttled/v2/store/memstore"

import (
	"github.com/throttled/throttled/v2"
	"sync"
	"sync/atomic"
	"time"

	lru "github.com/hashicorp/golang-lru"
)

// MemStore is an in-memory store implementation for throttled. It
// supports evicting the least recently used keys to control memory
// usage. It is stored in memory in the current process and thus
// doesn't share state with other rate limiters.
type MemStore struct {
	sync.RWMutex
	keys    *lru.Cache
	m       map[string]*int64
	timeNow func() time.Time //usually time.Now, but can be overridden for unit tests
}

// New initializes a Store. If maxKeys > 0, the number of different
// keys is restricted to the specified amount. In this case, it uses
// an LRU algorithm to evict older keys to make room for newer
// ones. If maxKeys <= 0, there is no limit on the number of keys,
// which may use an unbounded amount of memory.
func New(maxKeys int) (*MemStore, error) {
	var m *MemStore

	if maxKeys > 0 {
		keys, err := lru.New(maxKeys)
		if err != nil {
			return nil, err
		}

		m = &MemStore{
			keys:    keys,
			timeNow: time.Now,
		}
	} else {
		m = &MemStore{
			m:       make(map[string]*int64),
			timeNow: time.Now,
		}
	}
	return m, nil
}

// NewCtx is the version of New that can be used with a context-aware ratelimiter.
func NewCtx(maxKeys int) (throttled.GCRAStoreCtx, error) {
	st, err := New(maxKeys)
	return throttled.WrapStoreWithContext(st), err
}

// SetTimeNow makes this store use the given function instead of time.Now().
// This is useful for unit tests that use a simulated wallclock.
func (ms *MemStore) SetTimeNow(timeNow func() time.Time) {
	ms.timeNow = timeNow
}

// GetWithTime returns the value of the key if it is in the store or
// -1 if it does not exist. It also returns the current local time on
// the machine.
func (ms *MemStore) GetWithTime(key string) (int64, time.Time, error) {
	now := ms.timeNow()
	valP, ok := ms.get(key, false)

	if !ok {
		return -1, now, nil
	}

	return atomic.LoadInt64(valP), now, nil
}

// SetIfNotExistsWithTTL sets the value of key only if it is not
// already set in the store it returns whether a new value was set. It
// ignores the ttl.
func (ms *MemStore) SetIfNotExistsWithTTL(key string, value int64, _ time.Duration) (bool, error) {
	_, ok := ms.get(key, false)

	if ok {
		return false, nil
	}

	ms.Lock()
	defer ms.Unlock()

	_, ok = ms.get(key, true)

	if ok {
		return false, nil
	}

	// Store a pointer to a new instance so that the caller
	// can't mutate the value after setting
	v := value

	if ms.keys != nil {
		ms.keys.Add(key, &v)
	} else {
		ms.m[key] = &v
	}

	return true, nil
}

// CompareAndSwapWithTTL atomically compares the value at key to the
// old value. If it matches, it sets it to the new value and returns
// true. Otherwise, it returns false. If the key does not exist in the
// store, it returns false with no error. It ignores the ttl.
func (ms *MemStore) CompareAndSwapWithTTL(key string, old, new int64, _ time.Duration) (bool, error) {
	valP, ok := ms.get(key, false)

	if !ok {
		return false, nil
	}

	return atomic.CompareAndSwapInt64(valP, old, new), nil
}

func (ms *MemStore) get(key string, locked bool) (*int64, bool) {
	var valP *int64
	var ok bool

	if ms.keys != nil {
		var valI interface{}

		valI, ok = ms.keys.Get(key)
		if ok {
			valP = valI.(*int64)
		}
	} else {
		if !locked {
			ms.RLock()
			defer ms.RUnlock()
		}
		valP, ok = ms.m[key]
	}

	return valP, ok
}
