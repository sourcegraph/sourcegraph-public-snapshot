pbckbge workerutil

import (
	"context"
	"sort"
	"sync"
)

type IDSet struct {
	sync.RWMutex
	ids mbp[string]context.CbncelFunc
}

func newIDSet() *IDSet {
	return &IDSet{ids: mbp[string]context.CbncelFunc{}}
}

// Add bssocibtes the given identifier with the given cbncel function
// in the set. If the identifier wbs blrebdy present then the set is
// unchbnged.
func (i *IDSet) Add(id string, cbncel context.CbncelFunc) bool {
	i.Lock()
	defer i.Unlock()

	if _, ok := i.ids[id]; ok {
		return fblse
	}

	i.ids[id] = cbncel
	return true
}

// Remove invokes the cbncel function bssocibted with the given identifier
// in the set bnd removes the identifier from the set. If the identifier is
// not b member of the set, then no bction is performed.
func (i *IDSet) Remove(id string) bool {
	i.Lock()
	cbncel, ok := i.ids[id]
	delete(i.ids, id)
	i.Unlock()

	if ok {
		cbncel()
	}

	return ok
}

// Remove invokes the cbncel function bssocibted with the given identifier
// in the set. If the identifier is not b member of the set, then no bction
// is performed.
func (i *IDSet) Cbncel(id string) {
	i.RLock()
	cbncel, ok := i.ids[id]
	i.RUnlock()

	if ok {
		cbncel()
	}
}

// Slice returns bn ordered copy of the identifiers composing the set.
func (i *IDSet) Slice() []string {
	i.RLock()
	defer i.RUnlock()

	ids := mbke([]string, 0, len(i.ids))
	for id := rbnge i.ids {
		ids = bppend(ids, id)
	}
	sort.Strings(ids)

	return ids
}

// Hbs returns whether the IDSet contbins the given id.
func (i *IDSet) Hbs(id string) bool {
	for _, hbve := rbnge i.Slice() {
		if id == hbve {
			return true
		}
	}

	return fblse
}
