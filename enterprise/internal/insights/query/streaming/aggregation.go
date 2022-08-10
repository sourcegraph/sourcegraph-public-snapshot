package streaming

import (
	"container/heap"
	"sort"
	"strings"
)

type Aggregate struct {
	value string
	count int32
}

func (a *Aggregate) Less(b *Aggregate) bool {
	if a.count == b.count {
		// Sort alphabetically if of same count.
		return strings.Compare(a.value, b.value) < 0
	}
	return a.count > b.count
}

type aggregated map[string]*Aggregate

func (a aggregated) Add(value string, count int32) {
	result, ok := a[value]
	if !ok {
		a[value] = &Aggregate{value, count}
	} else {
		result.count += count
	}
}

// SortAggregate returns an ordered slice of elements to present to the user.
func (a aggregated) SortAggregate(max int) []*Aggregate {
	heap := aggregateHeap{max: max}
	for _, elt := range a {
		heap.Add(elt)
	}
	sort.Sort(heap.aggregateSlice)

	return heap.aggregateSlice
}

type aggregateSlice []*Aggregate

func (as aggregateSlice) Len() int {
	return len(as)
}

func (as aggregateSlice) Less(i, j int) bool {
	return as[i].Less(as[j])
}

func (as aggregateSlice) Swap(i, j int) {
	as[i], as[j] = as[j], as[i]
}

type aggregateHeap struct {
	aggregateSlice
	max int
}

func (ah *aggregateHeap) Add(a *Aggregate) {
	if len(ah.aggregateSlice) < ah.max {
		heap.Push(ah, a)
	} else if ah.max > 0 && a.Less(ah.aggregateSlice[0]) {
		// We keep the element with the least count at index 0.
		heap.Pop(ah)
		heap.Push(ah, a)
	}
}

func (ah *aggregateHeap) Less(i, j int) bool {
	// We want a min heap i.e. the head of the heap is the least important value we have kept so far.
	return ah.aggregateSlice[j].Less(ah.aggregateSlice[i])
}

func (ah *aggregateHeap) Push(a any) {
	ah.aggregateSlice = append(ah.aggregateSlice, a.(*Aggregate))
}

func (ah *aggregateHeap) Pop() any {
	old := ah.aggregateSlice
	n := len(old)
	a := old[n-1]
	ah.aggregateSlice = old[0 : n-1]
	return a
}
