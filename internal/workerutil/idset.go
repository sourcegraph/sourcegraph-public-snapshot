package workerutil

import (
	"context"
	"sync"
)

type internalWrapper[T Record] struct {
	cancel context.CancelFunc
	r      T
}

type RecordSet[T Record] struct {
	sync.RWMutex
	records map[int]internalWrapper[T]
}

func newIDSet[T Record]() *RecordSet[T] {
	return &RecordSet[T]{records: map[int]internalWrapper[T]{}}
}

// Add associates the given identifier with the given cancel function
// in the set. If the identifier was already present then the set is
// unchanged.
func (i *RecordSet[T]) Add(r T, cancel context.CancelFunc) bool {
	i.Lock()
	defer i.Unlock()

	if _, ok := i.records[r.RecordID()]; ok {
		return false
	}

	i.records[r.RecordID()] = internalWrapper[T]{
		cancel: cancel,
		r:      r,
	}
	return true
}

// Remove invokes the cancel function associated with the given identifier
// in the set and removes the identifier from the set. If the identifier is
// not a member of the set, then no action is performed.
func (i *RecordSet[T]) Remove(r T) bool {
	i.Lock()
	w, ok := i.records[r.RecordID()]
	delete(i.records, r.RecordID())
	i.Unlock()

	if ok {
		w.cancel()
	}

	return ok
}

// Remove invokes the cancel function associated with the given identifier
// in the set. If the identifier is not a member of the set, then no action
// is performed.
func (i *RecordSet[T]) Cancel(r T) {
	i.RLock()
	w, ok := i.records[r.RecordID()]
	i.RUnlock()

	if ok {
		w.cancel()
	}
}

// Slice returns an ordered copy of the identifiers composing the set.
func (i *RecordSet[T]) Slice() []T {
	i.RLock()
	defer i.RUnlock()

	ids := make([]T, 0, len(i.records))
	for _, r := range i.records {
		ids = append(ids, r.r)
	}
	// TODO: This is likely only for testing, but should still reimplement it.
	// sort.Ints(ids)

	return ids
}

// Has returns whether the IDSet contains the given id.
func (i *RecordSet[T]) Has(r T) bool {
	id := r.RecordID()
	for _, have := range i.Slice() {
		if id == have.RecordID() {
			return true
		}
	}

	return false
}
