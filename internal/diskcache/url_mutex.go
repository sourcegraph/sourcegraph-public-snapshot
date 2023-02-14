package diskcache

import (
	"sync"
	"sync/atomic"
)

// If we're saving to the local FS, we need to globally synchronize
// writes so we don't corrupt the .zip files with concurrent
// writes. We also needn't bother fetching the same file concurrently,
// since we'll be able to reuse it in the second caller.

var (
	urlMusMu sync.Mutex
	urlMus   = map[string]*queryableMutex{}
)

func urlMu(path string) *queryableMutex {
	urlMusMu.Lock()
	mu, ok := urlMus[path]
	if !ok {
		mu = new(queryableMutex)
		urlMus[path] = mu
	}
	urlMusMu.Unlock()
	return mu
}

type queryableMutex struct {
	mu     sync.Mutex
	locked atomic.Int64
}

// Lock is a wrapper around Mutex.Lock.
func (qm *queryableMutex) Lock() {
	// Order is important. If we increment locked after mu.Lock then there
	// will be a period where locked is zero. We want calls to IsLocked to
	// conservatively return true.
	qm.locked.Add(1)
	qm.mu.Lock()
}

// Lock is a wrapper around Mutex.Unlock.
func (qm *queryableMutex) Unlock() {
	// Order isn't as important, but in the same spirit we would rather return
	// true in IsLocked in case another goroutine comes in to lock soon.
	qm.mu.Unlock()
	qm.locked.Add(-1)
}

// IsLocked returns true if a goroutine currently holds the lock or is about
// to / has just released the lock. This shouldn't be used to implement
// synchronisation logic, but should be used to advise for reporting back to
// users.
func (qm *queryableMutex) IsLocked() bool {
	return qm.locked.Load() > 0
}
