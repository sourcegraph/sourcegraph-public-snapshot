package golang

import "sync"

var namedLocksMu sync.Mutex
var namedLocks = map[string]*sync.Mutex{}

func lock(s string) func() {
	namedLocksMu.Lock()
	mu := namedLocks[s]
	if mu == nil {
		mu = &sync.Mutex{}
		namedLocks[s] = mu
	}
	namedLocksMu.Unlock()

	mu.Lock()
	return func() {
		mu.Unlock()
	}
}
