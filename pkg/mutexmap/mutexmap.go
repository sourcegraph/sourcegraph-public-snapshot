// Package mutexmap provides a lightweight way to synchronize access over
// named resources
package mutexmap

import "sync"

// MutexMap provides methods to lock named resources. Use New
type MutexMap struct {
	mu sync.Mutex
	m  map[string]*sync.Mutex
}

// New creates a new MutexMap
func New() *MutexMap {
	return &MutexMap{
		m: map[string]*sync.Mutex{},
	}
}

// Lock is like sync.Mutex.Lock, but for a specific key
func (s *MutexMap) Lock(k string) {
	s.getMu(k).Lock()
}

// Unlock is like sync.Mutex.Unlock, but for a specific key
func (s *MutexMap) Unlock(k string) {
	s.getMu(k).Unlock()
}

func (s *MutexMap) getMu(k string) *sync.Mutex {
	s.mu.Lock()
	mu, ok := s.m[k]
	if !ok {
		mu = &sync.Mutex{}
		s.m[k] = mu
	}
	s.mu.Unlock()
	return mu
}
