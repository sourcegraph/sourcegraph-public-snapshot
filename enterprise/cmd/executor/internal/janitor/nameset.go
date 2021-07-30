package janitor

import (
	"sort"
	"sync"
)

type NameSet struct {
	sync.RWMutex
	names map[string]struct{}
}

func NewNameSet() *NameSet {
	return &NameSet{names: map[string]struct{}{}}
}

func (s *NameSet) Add(name string) {
	s.Lock()
	s.names[name] = struct{}{}
	s.Unlock()
}

func (s *NameSet) Remove(name string) {
	s.Lock()
	delete(s.names, name)
	s.Unlock()
}

func (s *NameSet) Slice() []string {
	s.RLock()
	defer s.RUnlock()

	names := make([]string, 0, len(s.names))
	for name := range s.names {
		names = append(names, name)
	}
	sort.Strings(names)

	return names
}
