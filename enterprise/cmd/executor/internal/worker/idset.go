package worker

import (
	"context"
	"sort"
	"sync"
)

type IDSet struct {
	sync.RWMutex
	ids map[int]context.CancelFunc
}

func newIDSet() *IDSet {
	return &IDSet{ids: map[int]context.CancelFunc{}}
}

// Add associates the given identifier with the given cancel function
// in the set. If the identifier was already present then the set is
// unchanged.
func (i *IDSet) Add(id int, cancel context.CancelFunc) bool {
	i.Lock()
	defer i.Unlock()

	if _, ok := i.ids[id]; ok {
		return false
	}

	i.ids[id] = cancel
	return true
}

// Remove invokes the cancel function associated with the given identifier
// in the setand removes the identifier from the set. If the identifier is
// not a member of the set, then no action is performed.
func (i *IDSet) Remove(id int) {
	i.Lock()
	cancel, ok := i.ids[id]
	delete(i.ids, id)
	i.Unlock()

	if ok {
		cancel()
	}
}

// Slice returns an ordered copy of the identifiers composing the set.
func (i *IDSet) Slice() []int {
	i.RLock()
	defer i.RUnlock()

	ids := make([]int, 0, len(i.ids))
	for id := range i.ids {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	return ids
}
