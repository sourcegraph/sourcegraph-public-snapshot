package apiworker

import (
	"sort"
	"sync"
)

type IDSet struct {
	sync.RWMutex
	ids map[int]struct{}
}

func newIDSet() *IDSet {
	return &IDSet{ids: map[int]struct{}{}}
}

func (i *IDSet) Add(id int) {
	i.Lock()
	i.ids[id] = struct{}{}
	i.Unlock()
}

func (i *IDSet) Remove(id int) {
	i.Lock()
	delete(i.ids, id)
	i.Unlock()
}

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
