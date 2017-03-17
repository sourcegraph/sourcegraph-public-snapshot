package mutexmap

import "sync"

type MutexMap struct {
	mu sync.Mutex
	v  map[string]*sync.Mutex
}

func (m *MutexMap) Get(key string) *sync.Mutex {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.v == nil {
		m.v = map[string]*sync.Mutex{}
	}
	v, ok := m.v[key]
	if !ok {
		v = new(sync.Mutex)
		m.v[key] = v
	}
	return v
}
