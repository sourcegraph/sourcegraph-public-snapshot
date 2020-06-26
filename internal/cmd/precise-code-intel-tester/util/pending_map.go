package util

import (
	"sort"
	"sync"
)

type pendingMap struct {
	sync.RWMutex
	pending map[int]bool
}

// newPendingMap creates a new pending map with n pending tasks.
func newPendingMap(n int) *pendingMap {
	pending := make(map[int]bool, n)
	for i := 0; i < n; i++ {
		pending[i] = false
	}

	return &pendingMap{pending: pending}
}

func (m *pendingMap) remove(i int) {
	m.Lock()
	defer m.Unlock()
	delete(m.pending, i)
}

func (m *pendingMap) keys() (keys []int) {
	m.RLock()
	defer m.RUnlock()

	for k := range m.pending {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

func (m *pendingMap) set(i int) {
	m.Lock()
	defer m.Unlock()
	m.pending[i] = true
}

func (m *pendingMap) get(i int) bool {
	m.RLock()
	defer m.RUnlock()
	return m.pending[i]
}

func (m *pendingMap) size() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.pending)
}
