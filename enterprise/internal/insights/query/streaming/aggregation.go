package streaming

import (
	"sort"
	"strings"
)

type aggregated struct {
	maxResults int
	Results    map[string]int32
	Overflow   int32
}

type Aggregate struct {
	Value string
	Count int32
}

func (a *aggregated) Add(value string, count int32) {
	if _, ok := a.Results[value]; !ok {
		if len(a.Results) >= a.maxResults {
			a.Overflow += count
		} else {
			a.Results[value] = count
		}
	} else {
		a.Results[value] += count
	}
}

func (a aggregated) SortAggregate() []*Aggregate {
	aggregateSlice := make(aggregateSlice, 0, len(a.Results))
	for val, count := range a.Results {
		aggregateSlice = append(aggregateSlice, &Aggregate{val, count})
	}
	sort.Sort(aggregateSlice)

	return aggregateSlice
}

type aggregateSlice []*Aggregate

func (a *Aggregate) Less(b *Aggregate) bool {
	if a.Count == b.Count {
		// Sort alphabetically if of same count.
		return strings.Compare(a.Value, b.Value) < 0
	}
	return a.Count > b.Count
}

func (as aggregateSlice) Len() int {
	return len(as)
}

func (as aggregateSlice) Less(i, j int) bool {
	return as[i].Less(as[j])
}

func (as aggregateSlice) Swap(i, j int) {
	as[i], as[j] = as[j], as[i]
}
